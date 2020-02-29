package i18n

import (
	"net/http"

	"github.com/BurntSushi/toml"
	gi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"gopkg.in/fsnotify.v1"
)

var bundle *gi18n.Bundle

func init() {
	bundle = gi18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	w, _ := fsnotify.NewWatcher()
	err := w.Add("i18n/translations")
	if err != nil {
		log.Error().Err(err).Msg("Can't auto reload translations")
	} else {
		go func() {
			for {
				<-w.Events
				log.Info().Msg("Reloading Translations")
				bundle.MustLoadMessageFile("i18n/translations/active.en.toml")
				bundle.MustLoadMessageFile("i18n/translations/active.sv.toml")
				bundle.MustLoadMessageFile("i18n/translations/active.fr.toml")

			}
		}()
	}
	bundle.MustLoadMessageFile("i18n/translations/active.en.toml")
	bundle.MustLoadMessageFile("i18n/translations/active.sv.toml")
	bundle.MustLoadMessageFile("i18n/translations/active.fr.toml")

}

type Translator struct {
	l *gi18n.Localizer
}

func (t *Translator) localizeFromConfig(c *gi18n.LocalizeConfig) string {
	s, err := t.l.Localize(c)
	if err != nil {
		panic(err)
	}
	return s
}

func (t *Translator) Localize(s string) string {
	return t.localizeFromConfig(&gi18n.LocalizeConfig{MessageID: s})
}

func (t *Translator) LocalizeWithData(s string, d map[string]interface{}) string {
	return t.localizeFromConfig(&gi18n.LocalizeConfig{MessageID: s, TemplateData: d})
}

func GetTranslator(r *http.Request) *Translator {
	accept := r.Header.Get("Accept-Language")
	lang := r.FormValue("lang")
	localizer := gi18n.NewLocalizer(bundle, lang, accept)
	return &Translator{
		l: localizer,
	}
}
