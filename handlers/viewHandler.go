package handlers

import (
	"fmt"
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
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	vars := mux.Vars(r)
	user, loggedIn := DB.CheckForLoggedInUser(r)

	postID, _ := strconv.Atoi(vars["postID"])
	post, ok := DB.Post(int64(postID))
	if !ok {
		return
	}
	fmt.Println(post)

	poster, _ := DB.User(post.Poster)

	templateInfo := ViewResultsTemplate{
		Post:         post,
		Author:       poster,
		IsAbleToEdit: (user.Admin || post.Poster == user.Username) && loggedIn,
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
