package main

import (
	"github.com/gorilla/mux"
	"github.com/pytimer/mux-logrus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
)

var DB DBType

func cacheMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=2592000")
    next.ServeHTTP(w, r)
  })
}

func main() {
	log.Info("starting, loading db")
	DB = LoadDB()
	log.Info("db loaded")

	r := mux.NewRouter()
	r.Use(muxlogrus.NewLogger().Middleware)
	r.HandleFunc("/", rootHandler)
	r.HandleFunc("/rules", rulesHandler)

	r.HandleFunc("/setup", setupPageHandler).Methods("GET")
	r.HandleFunc("/setup", setupHandler).Methods("POST")
	r.HandleFunc("/deletePost/{postID}", deletePostPageHandler).Methods("GET")
	r.HandleFunc("/deletePost/{postID}", deletePostHandler).Methods("POST")
	r.HandleFunc("/register", registerPageHandler).Methods("GET")
	r.HandleFunc("/register", registerHandler).Methods("POST")
	r.HandleFunc("/search", searchHandler)
	r.HandleFunc("/login", loginPageHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/upload", uploadHandler).Methods("POST")
	r.HandleFunc("/upload", uploadPageHandler).Methods("GET")
	r.HandleFunc("/editPost/{postID}", editPostHandler).Methods("POST")
	r.HandleFunc("/editUser/{userID}", editUserHandler).Methods("POST")

	r.HandleFunc("/view/{postID}", viewHandler)
	r.HandleFunc("/user/{userID}", userHandler)

	r.PathPrefix("/content/").Handler(cacheMiddleware(http.StripPrefix("/content/", http.FileServer(http.Dir("content")))))
	r.PathPrefix("/css/").Handler(cacheMiddleware(http.StripPrefix("/css/", http.FileServer(http.Dir("css")))))
	r.PathPrefix("/js/").Handler(cacheMiddleware(http.StripPrefix("/js/", http.FileServer(http.Dir("js")))))
	r.HandleFunc("/thumbnail/{postID}", thumbnailHandler)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go http.ListenAndServe("0.0.0.0:9090", r)
	<-c
	DB.Save()
	log.Info("Exiting")

}
