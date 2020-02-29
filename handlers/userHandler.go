package handlers

import (
	"net/http"

	"github.com/NamedKitten/kittehimageboard/i18n"
	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/gorilla/mux"
)

type UserResultsTemplate struct {
	AvatarPost   types.Post
	User         types.User
	IsAbleToEdit bool
	templates.T
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	vars := mux.Vars(r)
	loggedInUser, loggedIn := DB.CheckForLoggedInUser(r)

	username := vars["userID"]
	user, exist := DB.User(vars["userID"])
	if !exist {
		return
	}

	avatarPost, _ := DB.Post(user.AvatarID)
	templateInfo := UserResultsTemplate{
		AvatarPost:   avatarPost,
		User:         user,
		IsAbleToEdit: (loggedInUser.Username == username) && loggedIn,
		T: templates.T{
			LoggedIn:     loggedIn,
			LoggedInUser: loggedInUser,
			Translator:   i18n.GetTranslator(r),
		},
	}

	err := templates.RenderTemplate(w, "user.html", templateInfo)
	if err != nil {
		panic(err)
	}
}
