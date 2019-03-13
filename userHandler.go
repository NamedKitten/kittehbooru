package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type UserResultsTemplate struct {
	User       User
	IsAbleToEdit bool
	TemplateTemplate
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loggedInUser, loggedIn := DB.CheckForLoggedInUser(r)

	userID, _ := strconv.Atoi(vars["userID"])
	var user User
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

	templateInfo := UserResultsTemplate {
		User:       user,
		IsAbleToEdit: (loggedInUser.ID == int64(userID)) && loggedIn,
		TemplateTemplate: TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: loggedInUser,
		},
	}

	err := renderTemplate(w, "user.html", templateInfo)
	if err != nil {
		panic(err)
	}
}
