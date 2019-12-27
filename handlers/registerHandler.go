package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/NamedKitten/kittehimageboard/utils"
	"github.com/rs/zerolog/log"
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
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", 302)
		return
	}
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if !DB.VerifyRecaptcha(r.FormValue("recaptchaResponse")) {
		http.Redirect(w, r, "/register", 302)
		return
	}

	_, err := DB.User(username)
	if err != nil {
		http.Redirect(w, r, "/register", 302)
		return
	}

	DB.AddUser(types.User{
		Username:    username,
		Description: "",
		Posts:       []int64{},
	})
	DB.Passwords[username] = utils.EncryptPassword(password)
	log.Info().Str("username", username).Msg("Register")
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   DB.CreateSession(username),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", 302)
	return

}
