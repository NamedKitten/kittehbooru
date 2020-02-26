package handlers

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// EditUserHandler is the endpoint used to edit user settings.
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loggedInUser, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}

	user, exist := DB.User(vars["userID"])
	if !exist {
		http.Redirect(w, r, "/", 302)
		return
	}

	if !(user.Admin || loggedInUser.Username == user.Username) {
		http.Redirect(w, r, "/", 302)
		return
	}

	r.ParseForm()
	description := r.PostFormValue("description")
	if !(len(description) == 0) {
		user.Description = description
	}

	password := r.PostFormValue("password")
	if !(len(password) == 0) {
		DB.SetPassword(user.Username, password)
	}

	avatarID := r.PostFormValue("avatarID")
	if !(len(avatarID) == 0) {
		avatarIDInt, _ := strconv.Atoi(avatarID)
		user.AvatarID = int64(avatarIDInt)
	}

	DB.EditUser(user)

	http.Redirect(w, r, "/user/"+vars["userID"], 302)
}
