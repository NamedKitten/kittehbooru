//+build development

package start

import (
	_"net/http/pprof"
	"net/http"

	"github.com/gorilla/mux"
)

func addPprof(r *mux.Router) {
	r.PathPrefix("/debug/pprof").Handler(http.DefaultServeMux)
}
