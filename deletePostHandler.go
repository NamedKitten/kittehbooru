package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

type DeletePostTemplate struct {
	PostID int64
	TemplateTemplate
}

func deletePostPageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, _ := strconv.Atoi(vars["postID"])
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}


	templateInfo := DeletePostTemplate{
		PostID:         int64(postID),
		TemplateTemplate: TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}

	err := renderTemplate(w, "deletePost.html", templateInfo)
	if err != nil {
		panic(err)
	}
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		log.Error("Not logged in.")
		http.Redirect(w, r, "/login", 302)
		return
	}

	postID, _ := strconv.Atoi(vars["postID"])

	post, postExists := DB.Posts[int64(postID)]
	if !postExists {
		log.Error("Post doesn't exist.")
		http.Redirect(w, r, "/", 302)
		return
	}

	if !(user.Owner || user.Admin || post.PosterID == user.ID) {
		log.Error("Not authorized.")
		http.Redirect(w, r, "/", 302)
		return
	}
	log.Error("Deleting.")

	DB.DeletePost(int64(postID))

	http.Redirect(w, r, "/", 302)
}
