//+build development

package start

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/mux"
)

func addPprof(r *mux.Router) {
	r.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)
}
