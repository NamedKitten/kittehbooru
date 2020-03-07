//+build !development

package start

import (
	"github.com/gorilla/mux"
)

func addPprof(*mux.Router) {
}
