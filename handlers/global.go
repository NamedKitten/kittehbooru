package handlers

import (
	"net/http"

	"github.com/NamedKitten/kittehimageboard/database"
	"github.com/rs/zerolog/log"
)

var DB *database.DB

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)

	_, err := w.Write([]byte(message))
	if err != nil {
		log.Error().Err(err).Msg("RenderError Error")
	}
}
