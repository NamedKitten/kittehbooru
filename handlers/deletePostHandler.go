package handlers

import (
	"net/http"
	"strconv"

	"github.com/NamedKitten/kittehimageboard/i18n"
	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type DeletePostTemplate struct {
	PostID int64
	templates.T
}

func DeletePostPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, _ := strconv.Atoi(vars["postID"])
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
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

	err := templates.RenderTemplate(w, "deletePost.html", templateInfo)
	if err != nil {
		panic(err)
	}
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	postID, _ := strconv.Atoi(vars["postID"])

	post, postExists := DB.Post(int64(postID))
	if !postExists {
		http.Redirect(w, r, "/", 302)
		return
	}

	if !(user.Owner || user.Admin || post.Poster == user.Username) {
		http.Redirect(w, r, "/", 302)
		return
	}

	err := DB.DeletePost(int64(postID))
	if err != nil {
		log.Error().Err(err).Msg("Delete Post")
		renderError(w, "DELETE_POST_ERR", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", 302)
}
