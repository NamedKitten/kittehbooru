package start

import (
	"net/http"
	"os"
	"os/signal"
	"runtime/trace"

	"github.com/NamedKitten/kittehbooru/database"
	"github.com/NamedKitten/kittehbooru/handlers"
	templates "github.com/NamedKitten/kittehbooru/template"

	//gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func taskMiddleware(name string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the http.Request Context with trace.Task object.
		taskCtx, task := trace.NewTask(r.Context(), name)
		defer task.End()
		log.Debug().Str("path", r.URL.Path).Str("method", r.Method).Msg("HTTP")

		r = r.WithContext(taskCtx)

		next.ServeHTTP(w, r)
	})
}

func Start(configFile string) {

	log.Info().Msg("Starting")
	DB = database.LoadDB(configFile)
	templates.DB = DB
	handlers.DB = DB
	log.Info().Msg("Loaded DB")

	r := mux.NewRouter()

	handleFunc := func(path string, f func(w http.ResponseWriter, r *http.Request)) *mux.Route {
		return r.Handle(path, taskMiddleware(path, http.HandlerFunc(f)))
	}
	handleFunc("/", handlers.RootHandler)
	handleFunc("/rules", handlers.RulesHandler)
	handleFunc("/setup", handlers.SetupPageHandler).Methods("GET")
	handleFunc("/setup", handlers.SetupHandler).Methods("POST")
	handleFunc("/deletePost/{postID}", handlers.DeletePostPageHandler).Methods("GET")
	handleFunc("/deletePost/{postID}", handlers.DeletePostHandler).Methods("POST")
	handleFunc("/deleteUser", handlers.DeleteUserPageHandler).Methods("GET")
	handleFunc("/deleteUser", handlers.DeleteUserHandler).Methods("POST")
	handleFunc("/register", handlers.RegisterPageHandler).Methods("GET")
	handleFunc("/register", handlers.RegisterHandler).Methods("POST")
	handleFunc("/search", handlers.SearchHandler)
	handleFunc("/login", handlers.LoginPageHandler).Methods("GET")
	handleFunc("/logout", handlers.LogoutHandler).Methods("GET")
	handleFunc("/login", handlers.LoginHandler).Methods("POST")
	handleFunc("/upload", handlers.UploadHandler).Methods("POST")
	handleFunc("/upload", handlers.UploadPageHandler).Methods("GET")
	handleFunc("/editPost/{postID}", handlers.EditPostHandler).Methods("POST")
	handleFunc("/editUser/{userID}", handlers.EditUserHandler).Methods("POST")
	handleFunc("/view/{postID}", handlers.ViewHandler)
	handleFunc("/user/{userID}", handlers.UserHandler)
	addPprof(r)

	r.PathPrefix("/content/").Handler(
		taskMiddleware("/content/",
			cacheMiddleware(
				http.StripPrefix("/content/",
					http.FileServer(DB.ContentStorage),
				))))
	r.PathPrefix("/css/").Handler(cacheMiddleware(http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css")))))
	r.PathPrefix("/js/").Handler(cacheMiddleware(http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js")))))
	handleFunc("/thumbnail/{postID}.webp", handlers.ThumbnailHandler)

	go func() {
		err := http.ListenAndServe(DB.Settings.ListenAddress, r)
		if err != nil {
			log.Error().Err(err).Msg("Can't start web")
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	DB.Save()
	log.Info().Msg("Exiting")
}
