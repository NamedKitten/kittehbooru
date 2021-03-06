package handlers

import (
	"net/http"
	"time"

	"github.com/NamedKitten/kittehbooru/i18n"
	templates "github.com/NamedKitten/kittehbooru/template"
	"github.com/NamedKitten/kittehbooru/utils"

	"github.com/NamedKitten/kittehbooru/types"
	"github.com/rs/zerolog/log"
)

func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.RenderTemplate(w, "register.html", templates.T{Translator: i18n.GetTranslator(r)})
	if err != nil {
		renderError(w, "TEMPLATE_RENDER_ERROR", err, http.StatusBadRequest)
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Parse Form")
		return
	}
	username := utils.FilterString(r.FormValue("username"))
	password := utils.FilterString(r.FormValue("password"))

	if (len(username) <= 0) || (len(password) <= 0) {
		http.Redirect(w, r, "/register", http.StatusFound)
	}

	if !DB.VerifyRecaptcha(ctx, r.FormValue("recaptchaResponse")) {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}

	_, err = DB.User(ctx, username)
	if err == nil {
		http.Redirect(w, r, "/register", http.StatusFound)
		return
	}
	u := types.User{
		Username:    username,
		Description: "",
		Posts:       []int64{},
	}
	DB.AddUser(ctx, u)

	err = DB.SetPassword(ctx, username, password)
	if err != nil {
		log.Error().Err(err).Msg("Register Password")
		return
	}
	log.Info().Str("username", username).Msg("Register")
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   DB.CreateSession(ctx, username),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
