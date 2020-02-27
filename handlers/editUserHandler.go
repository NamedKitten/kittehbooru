package handlers

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"github.com/rs/zerolog/log"
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

	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Parse Form")
		renderError(w, "PARSE_FORM_ERR", http.StatusBadRequest)
		return
	}
	description := r.PostFormValue("description")
	if !(len(description) == 0) {
		user.Description = description
	}

	password := r.PostFormValue("password")
	if !(len(password) == 0) {
		err = DB.SetPassword(user.Username, password)
		if err != nil {
			log.Error().Err(err).Msg("Set Password")
			renderError(w, "SET_PASSWORD_ERR", http.StatusBadRequest)
			return
		}
	}

	avatarID := r.PostFormValue("avatarID")
	if !(len(avatarID) == 0) {
		avatarIDInt, _ := strconv.Atoi(avatarID)
		user.AvatarID = int64(avatarIDInt)
	}

	err = DB.EditUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Edit User")
		renderError(w, "EDIT_USER_ERR", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/user/"+vars["userID"], 302)
}
