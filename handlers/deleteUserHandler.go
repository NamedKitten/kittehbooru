package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type DeleteUserTemplate struct {
	templates.TemplateTemplate
}

func DeleteUserPageHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	templateInfo := DeleteUserTemplate{
		TemplateTemplate: templates.TemplateTemplate{
			LoggedIn:     loggedIn,
			LoggedInUser: user,
		},
	}

	err := templates.RenderTemplate(w, "deleteUser.html", templateInfo)
	if err != nil {
		panic(err)
	}
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		log.Error("Not logged in.")
		http.Redirect(w, r, "/login", 302)
		return
	}

	if user.Owner {
		log.Error("The owner can't delete their account.")
		http.Redirect(w, r, "/", 302)
		return
	}
	log.WithFields(log.Fields{
		"username": user.Username,
	}).Info("Account Deletion")
	DB.DeleteUser(user.ID)

	http.Redirect(w, r, "/", 302)
}
