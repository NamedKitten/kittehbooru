package main

import (
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func registerPageHandler(w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(w, "register.html", nil)
	if err != nil {
		panic(err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if !DB.verifyRecaptcha(r.FormValue("recaptchaResponse")) {
		log.Error("Recaptcha failed: ", r.FormValue("recaptchaResponse"))
		http.Redirect(w, r, "/register", 302)
		return
	}

	_, usernameExists := DB.UsernameToID[username]
	if usernameExists {

		http.Redirect(w, r, "/register", 302)
		return
	}

	node, _ := snowflake.NewNode(1)
	userID := node.Generate().Int64()
	DB.Users[userID] = User{
		ID:          userID,
		Owner:       true,
		Admin:       true,
		Username:    username,
		Description: "",
		Posts:       []int64{},
	}
	DB.UsernameToID[username] = userID
	DB.Passwords[userID] = encryptPassword(password)

	sessionToken := genSessionToken()
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   sessionToken,
		Expires: time.Now().Add(3 * time.Hour),
	})
	DB.Sessions[sessionToken] = userID
	http.Redirect(w, r, "/", 302)
	return

}
