package handlers

import (
	"net/http"
)

// LogoutHandler is the endpoint used to log out.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	DB.Sessions.InvalidateSession(ctx, user.Username)
	http.Redirect(w, r, "/", http.StatusFound)
}
