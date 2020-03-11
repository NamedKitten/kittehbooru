package handlers

import (
	"net/http"

	"github.com/NamedKitten/kittehimageboard/types"

	"github.com/NamedKitten/kittehimageboard/i18n"
	templates "github.com/NamedKitten/kittehimageboard/template"
)

// RootTemplateData contains data to be used in the template.
// We require LoggedIn and User to display text and more buttons if
// a user is already logged in.
type RootTemplateData struct {
	PostPopularity []types.TagCounts
	templates.T
}

// rootHandler is the root endpoint where a index page is served.
func RootHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !DB.SetupCompleted {
		http.Redirect(w, r, "/setup", http.StatusFound)
		return
	}
	user, loggedIn := DB.CheckForLoggedInUser(ctx, r)
	x := RootTemplateData{
		T: templates.T{

			LoggedIn:     loggedIn,
			LoggedInUser: user,
			Translator:   i18n.GetTranslator(r),
		},
		PostPopularity: DB.TopNCommonTags(ctx, 20, []string{"*"}),
	}
	err := templates.RenderTemplate(w, "index.html", x)
	if err != nil {
		panic(err)
	}
}
