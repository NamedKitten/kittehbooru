package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// EditUserHandler is the endpoint used to edit user settings.
func EditUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	loggedInUser, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	user, exist := DB.User(ctx, vars["userID"])
	if !exist {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if !(user.Admin || loggedInUser.Username == user.Username) {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var err error
	err = r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Parse Form")
		return
	}
	description := r.PostFormValue("description")
	if !(len(description) == 0) {
		user.Description = description
	}

	password := r.PostFormValue("password")
	if !(len(password) == 0) {
		passwordPrevious := r.PostFormValue("passwordPrevious")
		log.Error().Msg(passwordPrevious)
		if !DB.CheckPassword(ctx, user.Username, passwordPrevious) {
			return
		}

		passwordConfirm := r.PostFormValue("passwordConfirm")
		if password != passwordConfirm {
			return
		}

		err = DB.SetPassword(ctx, user.Username, password)
		if err != nil {
			log.Error().Err(err).Msg("Set Password")
			return
		}
	}

	avatarID := r.PostFormValue("avatarID")
	if !(len(avatarID) == 0) {
		avatarIDInt, aerr := strconv.Atoi(avatarID)
		if aerr != nil {
			log.Error().Err(aerr).Msg("Can't convert avatarID to string")
			return
		}
		user.AvatarID = int64(avatarIDInt)
	}

	err = DB.EditUser(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Edit User")
		renderError(w, "EDIT_USER_ERR", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/user/"+vars["userID"], http.StatusFound)
}
