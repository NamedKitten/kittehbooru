package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type ViewResultsTemplate struct {
	Post         types.Post
	Author       types.User
	IsAbleToEdit bool
	templates.TemplateTemplate
}

func ViewHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(r)

	postID, _ := strconv.Atoi(vars["postID"])
	post, ok := DB.Posts[int64(postID)]
	if !ok {
		return
	}

	templateInfo := ViewResultsTemplate{
		Post:         post,
		Author:       DB.Users[post.PosterID],
		IsAbleToEdit: (user.Admin || post.PosterID == user.ID) && loggedIn,
		TemplateTemplate: templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}

	err := templates.RenderTemplate(w, "view.html", templateInfo)
	if err != nil {
		panic(err)
	}
}
