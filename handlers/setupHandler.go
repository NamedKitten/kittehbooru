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

// setupPageHandler takes you to the setup page for initial setup.
func SetupPageHandler(w http.ResponseWriter, r *http.Request) {
	if DB.SetupCompleted {
		http.Redirect(w, r, "/", 302)
		return
	}
	log.Error("Setup.")
	err := templates.RenderTemplate(w, "setup.html", nil)
	if err != nil {
		panic(err)
	}
}

func SetupHandler(w http.ResponseWriter, r *http.Request) {
	if DB.SetupCompleted == true {
		log.Error("Setup already completed. ", DB.SetupCompleted)
		http.Redirect(w, r, "/", 302)
		return
	}
	r.ParseForm()
	username := r.FormValue("adminUsername")
	password := r.FormValue("adminPassword")
	siteTitle := r.FormValue("siteTitle")
	rules := r.FormValue("rules")
	enablereCaptcha := r.FormValue("enablereCaptcha")
	reCaptchaPublicKey := r.FormValue("reCaptchaPublicKey")
	reCaptchaPrivateKey := r.FormValue("reCaptchaPrivateKey")
	log.Error("Stuff: ", username, password, siteTitle, enablereCaptcha, reCaptchaPublicKey, reCaptchaPrivateKey)
	log.Error(r.Form)
	reCaptchaEnabled := false
	if enablereCaptcha == "on" {
		reCaptchaEnabled = true
	}

	node, _ := snowflake.NewNode(1)
	userID := node.Generate().Int64()
	DB.Users[userID] = types.User{
		ID:          userID,
		Owner:       true,
		Admin:       true,
		Username:    username,
		Description: "",
		Posts:       []int64{},
	}
	DB.UsernameToID[username] = userID
	DB.Passwords[userID] = utils.EncryptPassword(password)

	DB.Settings.SiteName = siteTitle
	DB.Settings.Rules = rules
	DB.Settings.ReCaptcha = reCaptchaEnabled
	DB.Settings.ReCaptchaPubkey = reCaptchaPublicKey
	DB.Settings.ReCaptchaPrivkey = reCaptchaPrivateKey
	DB.SetupCompleted = true
	http.SetCookie(w, &http.Cookie{
		Name:    "sessionToken",
		Value:   DB.CreateSession(userID),
		Expires: time.Now().Add(3 * time.Hour),
	})
	http.Redirect(w, r, "/", 302)
}
