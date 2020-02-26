package handlers

import (
	"github.com/NamedKitten/kittehimageboard/template"
	"github.com/NamedKitten/kittehimageboard/types"
	"net/http"
	"time"
)

// setupPageHandler takes you to the setup page for initial setup.
func SetupPageHandler(w http.ResponseWriter, r *http.Request) {
	if DB.SetupCompleted {
		http.Redirect(w, r, "/", 302)
		return
	}
	err := templates.RenderTemplate(w, "setup.html", nil)
	if err != nil {
		panic(err)
	}
}

func formValueToBool(val string) bool {
	if val == "on" {
		return true
	}
	return false
}

func SetupHandler(w http.ResponseWriter, r *http.Request) {
	if DB.SetupCompleted == true {
		http.Redirect(w, r, "/", 302)
		return
	}
	r.ParseForm()
	username := r.FormValue("adminUsername")
	password := r.FormValue("adminPassword")
	siteTitle := r.FormValue("siteTitle")
	rules := r.FormValue("rules")
	reCaptchaPublicKey := r.FormValue("reCaptchaPublicKey")
	reCaptchaPrivateKey := r.FormValue("reCaptchaPrivateKey")

	DB.AddUser(types.User{
		Owner:       true,
		Admin:       true,
		Username:    username,
		Description: "",
		Posts:       []int64{},
	})
	DB.SetPassword(username, password)
	DB.Settings.ThumbnailFormat = r.FormValue("thumbnailType")
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
		Value:   DB.Sessions.CreateSession(username),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", 302)
}
