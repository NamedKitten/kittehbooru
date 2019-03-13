package main

import (
	"github.com/gorilla/mux"
	"github.com/pytimer/mux-logrus"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"github.com/NamedKitten/kittehimageboard/handlers"
	"github.com/NamedKitten/kittehimageboard/database"
	"github.com/NamedKitten/kittehimageboard/template"
)

var DB *database.DBType

func cacheMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "max-age=2592000")
    next.ServeHTTP(w, r)
  })
}

func main() {
	log.Info("starting, loading db")
	DB = database.LoadDB()
	templates.DB = DB
	handlers.DB = DB
	log.Info("db loaded")

	r := mux.NewRouter()
	r.Use(muxlogrus.NewLogger().Middleware)
	r.HandleFunc("/", handlers.RootHandler)
	r.HandleFunc("/rules", handlers.RulesHandler)

	r.HandleFunc("/setup", handlers.SetupPageHandler).Methods("GET")
	r.HandleFunc("/setup", handlers.SetupHandler).Methods("POST")
	r.HandleFunc("/deletePost/{postID}", handlers.DeletePostPageHandler).Methods("GET")
	r.HandleFunc("/deletePost/{postID}", handlers.DeletePostHandler).Methods("POST")
	r.HandleFunc("/register", handlers.RegisterPageHandler).Methods("GET")
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/search", handlers.SearchHandler)
	r.HandleFunc("/login", handlers.LoginPageHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/upload", handlers.UploadHandler).Methods("POST")
	r.HandleFunc("/upload", handlers.UploadPageHandler).Methods("GET")
	r.HandleFunc("/editPost/{postID}", handlers.EditPostHandler).Methods("POST")
	r.HandleFunc("/editUser/{userID}", handlers.EditUserHandler).Methods("POST")

	r.HandleFunc("/view/{postID}", handlers.ViewHandler)
	r.HandleFunc("/user/{userID}", handlers.UserHandler)

	r.PathPrefix("/content/").Handler(cacheMiddleware(http.StripPrefix("/content/", http.FileServer(http.Dir("content")))))
	r.PathPrefix("/css/").Handler(cacheMiddleware(http.StripPrefix("/css/", http.FileServer(http.Dir("css")))))
	r.PathPrefix("/js/").Handler(cacheMiddleware(http.StripPrefix("/js/", http.FileServer(http.Dir("js")))))
	r.HandleFunc("/thumbnail/{postID}", handlers.ThumbnailHandler)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go http.ListenAndServe("0.0.0.0:9090", r)
	<-c
	DB.Save()
	log.Info("Exiting")

}
