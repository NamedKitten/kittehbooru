package handlers

import (
	"net/http"
	"time"

	"github.com/NamedKitten/kittehimageboard/i18n"
	templates "github.com/NamedKitten/kittehimageboard/template"

	"github.com/NamedKitten/kittehimageboard/types"
	"github.com/rs/zerolog/log"
)

func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.RenderTemplate(w, "register.html", templates.T{Translator: i18n.GetTranslator(r)})
	if err != nil {
		panic(err)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Parse Form")
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	if (len(username) <= 0) || (len(password) <= 0) {
		http.Redirect(w, r, "/register", http.StatusFound)
	}

	if !DB.VerifyRecaptcha(r.FormValue("recaptchaResponse")) {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	_, exist := DB.User(username)
	if exist {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}
	u := types.User{
		Username:    username,
		Description: "",
		Posts:       []int64{},
	}
	DB.AddUser(u)

	err = DB.SetPassword(username, password)
	if err != nil {
		log.Error().Err(err).Msg("Register Password")
		return
	}
	log.Info().Str("username", username).Msg("Register")
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   DB.Sessions.CreateSession(username),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
