package handlers

import (
	"net/http"
	"strconv"

	"github.com/NamedKitten/kittehbooru/i18n"
	templates "github.com/NamedKitten/kittehbooru/template"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type DeletePostTemplate struct {
	PostID int64
	templates.T
}

func DeletePostPageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		log.Error().Err(err).Msg("Can't convert postID to string")
		return
	}
	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	templateInfo := DeletePostTemplate{
		PostID: int64(postID),
		T: templates.T{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
			Translator:   i18n.GetTranslator(r),
		},
	}

	err = templates.RenderTemplate(w, "deletePost.html", templateInfo)
	if err != nil {
		panic(err)
	}
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		log.Error().Err(err).Msg("Can't convert postID to string")
		return
	}
	post, err := DB.Post(ctx, int64(postID))
	if err != nil {
		log.Error().Err(err).Msg("Can't fetch post")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if !(user.Owner || user.Admin || post.Poster == user.Username) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err = DB.DeletePost(ctx, int64(postID))
	if err != nil {
		log.Error().Err(err).Msg("Delete Post")
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
