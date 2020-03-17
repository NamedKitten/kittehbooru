package handlers

import (
	"net/http"
	"time"

	"github.com/NamedKitten/kittehbooru/i18n"
	templates "github.com/NamedKitten/kittehbooru/template"
	"github.com/NamedKitten/kittehbooru/types"
	"github.com/rs/zerolog/log"
)

// setupPageHandler takes you to the setup page for initial setup.
func SetupPageHandler(w http.ResponseWriter, r *http.Request) {
	if DB.SetupCompleted {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	err := templates.RenderTemplate(w, "setup.html", templates.T{Translator: i18n.GetTranslator(r)})
	if err != nil {
		panic(err)
	}
}

func SetupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if DB.SetupCompleted {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Parse Form")
		return
	}
	username := r.FormValue("adminUsername")
	password := r.FormValue("adminPassword")
	rules := r.FormValue("rules")

	DB.AddUser(ctx, types.User{
		Owner:       true,
		Admin:       true,
		Username:    username,
		Description: "",
		Posts:       []int64{},
	})
	err = DB.SetPassword(ctx, username, password)
	if err != nil {
		log.Error().Err(err).Msg("Setup Password")
		return
	}
	DB.Settings.Rules = rules
	DB.SetupCompleted = true
	DB.Save()
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   DB.CreateSession(ctx, username),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
