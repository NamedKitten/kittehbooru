package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type DeletePostTemplate struct {
	PostID int64
	templates.TemplateTemplate
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
		TemplateTemplate: templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
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

	DB.DeletePost(int64(postID))

	http.Redirect(w, r, "/", 302)
}
