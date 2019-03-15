package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.RenderTemplate(w, "register.html", nil)
	if err != nil {
		panic(err)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if ! DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if !DB.VerifyRecaptcha(r.FormValue("recaptchaResponse")) {
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
	DB.Users[userID] = types.User{
		ID:          userID,
		Username:    username,
		Description: "",
		Posts:       []int64{},
	}
	DB.UsernameToID[username] = userID
	DB.Passwords[userID] = utils.EncryptPassword(password)

	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   DB.CreateSession(userID),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", 302)
	return

}
