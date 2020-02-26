package handlers

import (
	"net/http"
)

// LogoutHandler is the endpoint used to log out.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	user, loggedIn := DB.CheckForLoggedInUser(r)
	if !loggedIn {
		http.Redirect(w, r, "/login", 302)
		return
	}
	DB.Sessions.InvalidateSession(user.Username)
	http.Redirect(w, r, "/", 302)
}
