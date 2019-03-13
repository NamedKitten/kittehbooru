package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type ViewResultsTemplate struct {
	Post         Post
	Author       User
	IsAbleToEdit bool
	TemplateTemplate
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
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
		TemplateTemplate: TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}

	err := renderTemplate(w, "view.html", templateInfo)
	if err != nil {
		panic(err)
	}
}
