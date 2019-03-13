package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type UserResultsTemplate struct {
	User         types.User
	IsAbleToEdit bool
	templates.TemplateTemplate
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loggedInUser, loggedIn := DB.CheckForLoggedInUser(r)

	userID, _ := strconv.Atoi(vars["userID"])
	var user types.User
	var ok bool

	user, ok = DB.Users[int64(userID)]
	if !ok {
		// See if it is a username rather than ID.
		newUserID, ok := DB.UsernameToID[vars["userID"]]
		if !ok {
			return
		} else {
			userID = int(newUserID)
			user = DB.Users[int64(userID)]
		}
	}

	templateInfo := UserResultsTemplate{
		User:         user,
		IsAbleToEdit: (loggedInUser.ID == int64(userID)) && loggedIn,
		TemplateTemplate: templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: loggedInUser,
		},
	}

	err := templates.RenderTemplate(w, "user.html", templateInfo)
	if err != nil {
		panic(err)
	}
}
