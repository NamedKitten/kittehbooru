
package handlers

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"github.com/NamedKitten/kittehimageboard/utils"
)

// EditUserHandler is the endpoint used to edit user settings.
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loggedInUser, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		log.Error("Not logged in.")
		http.Redirect(w, r, "/login", 302)
		return
	}

	userID, _ := strconv.Atoi(vars["userID"])

	user, userExists := DB.Users[int64(userID)]
	if !userExists {
		log.Error("User doesn't exist.")
		http.Redirect(w, r, "/", 302)
		return
	}

	if !(user.Admin || loggedInUser.ID == user.ID) {
		log.Error("Not authorized.")
		http.Redirect(w, r, "/", 302)
		return
	}

	r.ParseForm()

	description := r.PostFormValue("description")
	if ! (len(description) == 0) {
		user.Description = r.PostFormValue("description")
	}

	password := r.PostFormValue("password")
	if ! (len(password) == 0) {
		DB.Passwords[int64(userID)] = utils.EncryptPassword(password)
	}

	avatarID := r.PostFormValue("avatarID")
	if ! (len(avatarID) == 0) {
		avatarIDInt, _ := strconv.Atoi(avatarID)
		user.AvatarID = int64(avatarIDInt)
	}

	DB.Users[int64(userID)] = user


	http.Redirect(w, r, "/user/"+vars["userID"], 302)
}

