package start

import (
	"github.com/NamedKitten/kittehimageboard/database"
	"github.com/NamedKitten/kittehimageboard/handlers"
	templates "github.com/NamedKitten/kittehimageboard/template"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
)

var DB *database.DB

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, immutable, max-age=2592000")
		next.ServeHTTP(w, r)
	})
}

func Start() {
	log.Info().Msg("Starting")
	DB = database.LoadDB()
	templates.DB = DB
	handlers.DB = DB
	log.Info().Msg("Loaded DB")

	r := mux.NewRouter()
	r.HandleFunc("/", handlers.RootHandler)
	r.HandleFunc("/rules", handlers.RulesHandler)
	r.HandleFunc("/setup", handlers.SetupPageHandler).Methods("GET")
	r.HandleFunc("/setup", handlers.SetupHandler).Methods("POST")
	r.HandleFunc("/deletePost/{postID}", handlers.DeletePostPageHandler).Methods("GET")
	r.HandleFunc("/deletePost/{postID}", handlers.DeletePostHandler).Methods("POST")
	r.HandleFunc("/deleteUser", handlers.DeleteUserPageHandler).Methods("GET")
	r.HandleFunc("/deleteUser", handlers.DeleteUserHandler).Methods("POST")
	r.HandleFunc("/register", handlers.RegisterPageHandler).Methods("GET")
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/search", handlers.SearchHandler)
	r.HandleFunc("/login", handlers.LoginPageHandler).Methods("GET")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/upload", handlers.UploadHandler).Methods("POST")
	r.HandleFunc("/upload", handlers.UploadPageHandler).Methods("GET")
	r.HandleFunc("/editPost/{postID}", handlers.EditPostHandler).Methods("POST")
	r.HandleFunc("/editUser/{userID}", handlers.EditUserHandler).Methods("POST")
	r.HandleFunc("/view/{postID}", handlers.ViewHandler)
	r.HandleFunc("/user/{userID}", handlers.UserHandler)

	r.PathPrefix("/content/").Handler(cacheMiddleware(http.StripPrefix("/content/", http.FileServer(DB.ContentStorage))))
	r.PathPrefix("/css/").Handler(cacheMiddleware(http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css")))))
	r.PathPrefix("/js/").Handler(cacheMiddleware(http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js")))))
	r.HandleFunc("/thumbnail/{postID}-{size}.{ext}", handlers.ThumbnailHandler)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		err := http.ListenAndServe("0.0.0.0:80", gorillaHandlers.LoggingHandler(os.Stdout, r))
		if err != nil {
			log.Error().Err(err).Msg("Can't start web")
			panic(err)
		}
	}()
	<-c
	DB.Save()
	log.Info().Msg("Exiting")
}
