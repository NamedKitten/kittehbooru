package main

import (
	"github.com/bwmarrin/snowflake"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// setupPageHandler takes you to the setup page for initial setup.
func setupPageHandler(w http.ResponseWriter, r *http.Request) {
	/*if DB.SetupCompleted == true {
		log.Error("Setup already completed. ", DB.SetupCompleted)
		http.Redirect(w, r, "/", 302)
		return
	}*/
	log.Error("Setup.")
	err := renderTemplate(w, "setup.html", nil)
	if err != nil {
		panic(err)
	}
}

func setupHandler(w http.ResponseWriter, r *http.Request) {
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

	DB.Settings.SiteName = siteTitle
	DB.Settings.Rules = rules
	DB.Settings.ReCaptcha = reCaptchaEnabled
	DB.Settings.ReCaptchaPubkey = reCaptchaPublicKey
	DB.Settings.ReCaptchaPrivkey = reCaptchaPrivateKey
	DB.SetupCompleted = true

}
