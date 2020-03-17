package handlers

import (
	"net/http"
	"strconv"

	"github.com/NamedKitten/kittehbooru/i18n"
	templates "github.com/NamedKitten/kittehbooru/template"
	"github.com/NamedKitten/kittehbooru/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type ViewResultsTemplate struct {
	Post         types.Post
	Author       types.User
	IsAbleToEdit bool
	templates.T
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)

	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		log.Error().Err(err).Msg("Can't parse postID")
		return
	}
	post, err := DB.Post(ctx, int64(postID))
	if err != nil {
		return
	}

	poster, _ := DB.User(ctx, post.Poster)

	templateInfo := ViewResultsTemplate{
		Post:         post,
		Author:       poster,
		IsAbleToEdit: (user.Admin || post.Poster == user.Username) && loggedIn,
		T: templates.T{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
			Translator:   i18n.GetTranslator(r),
		},
	}

	err = templates.RenderTemplate(w, "view.html", templateInfo)
	if err != nil {
		panic(err)
	}
}
