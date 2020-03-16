package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NamedKitten/kittehimageboard/database"
	templates "github.com/NamedKitten/kittehimageboard/template"
	"github.com/rs/zerolog/log"
)

var DB *database.DB

var NoPermissionsError = errors.New("No Permissions")

func renderError(w http.ResponseWriter, message string, e error, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	err := templates.RenderTemplate(w, "error.html", message+": "+fmt.Sprintln(e))

	if err != nil {
		log.Error().Err(err).Msg("RenderError Error")
	}
}
