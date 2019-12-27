package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/gorilla/mux"
	"net/http"
)

type UserResultsTemplate struct {
	AvatarPost   types.Post
	User         types.User
	IsAbleToEdit bool
	templates.TemplateTemplate
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	vars := mux.Vars(r)
	loggedInUser, loggedIn := DB.CheckForLoggedInUser(r)

	username := vars["userID"]
	user, err := DB.User(username)
	if err != nil {
		return
	}

	templateInfo := UserResultsTemplate{
		AvatarPost:   DB.Posts[user.AvatarID],
		User:         user,
		IsAbleToEdit: (loggedInUser.Username == username) && loggedIn,
		TemplateTemplate: templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: loggedInUser,
		},
	}

	err = templates.RenderTemplate(w, "user.html", templateInfo)
	if err != nil {
		panic(err)
	}
}
