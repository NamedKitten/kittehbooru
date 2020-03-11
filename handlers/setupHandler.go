package handlers

import (
	"net/http"
	"time"

	"github.com/NamedKitten/kittehimageboard/i18n"
	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
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

func formValueToBool(val string) bool {
	return val == "on"
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
	siteTitle := r.FormValue("siteTitle")
	rules := r.FormValue("rules")
	reCaptchaPublicKey := r.FormValue("reCaptchaPublicKey")
	reCaptchaPrivateKey := r.FormValue("reCaptchaPrivateKey")

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

	DB.Settings.PDFView = formValueToBool(r.FormValue("enablePDFViewing"))
	DB.Settings.VideoThumbnails = formValueToBool(r.FormValue("enableVideoThumbnails"))
	DB.Settings.PDFThumbnails = formValueToBool(r.FormValue("enablePDFThumbnails"))

	DB.Settings.SiteName = siteTitle
	DB.Settings.Rules = rules
	DB.Settings.ReCaptcha = formValueToBool(r.FormValue("enablereCaptcha"))
	DB.Settings.ReCaptchaPubkey = reCaptchaPublicKey
	DB.Settings.ReCaptchaPrivkey = reCaptchaPrivateKey
	DB.SetupCompleted = true
	DB.Save()
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   DB.CreateSession(ctx, username),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
