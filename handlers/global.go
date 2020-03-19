package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/NamedKitten/kittehbooru/database"
	templates "github.com/NamedKitten/kittehbooru/template"
	"github.com/rs/zerolog/log"
)

var DB *database.DB

var NoPermissionsError = errors.New("No Permissions")

func renderError(w http.ResponseWriter, message string, e error, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)

	errorString := fmt.Sprintln(e)	
	err := templates.RenderTemplate(w, "error.html", message+": "+errorString)

	if err != nil {
		log.Error().Err(err).Msg("RenderError Error")
	}
}
