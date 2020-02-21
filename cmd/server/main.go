package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/meddion/web-blog/pkg/config"
	h "github.com/meddion/web-blog/pkg/handlers"
	"github.com/meddion/web-blog/pkg/models"
	log "github.com/sirupsen/logrus"
)

// In main we set up our endpoints (along with middleware)
// and start listening for upcoming requests
func main() {
	// Getting our config struct
	conf := config.GetConf()

	// Connecting to the DB
	if err := models.InitDB(conf.Db.URI, conf.Db.Name); err != nil {
		log.Fatal(err)
	}

	// Creating our router
	r := mux.NewRouter()

	// Setting up our session-auth middleware
	// Passing routes that do not require authorization to NewSessionAuthMiddleware
	signupEndpoint := "/signup/" + randSeq(32)
	sessionAuthMiddleware, err := h.NewSessionAuthMiddleware(
		"/api/account/login",
		"/api/account"+signupEndpoint,
		"/api/account/{name}",
		"/api/posts/info",
		"/api/posts/{pageNum:[0-9]+}",
		"/api/post/{id}",
	)
	if err != nil {
		log.Panic(err)
	}
	r.Use(sessionAuthMiddleware.Middleware)

	// Setting up endpoints with /api prefix in common
	api := r.PathPrefix("/api").Subrouter()
	accountRouter := api.PathPrefix("/account").Subrouter()
	accountRouter.HandleFunc("/login", h.LoginHandler).Methods("POST")
	accountRouter.HandleFunc("/logout", h.LogoutHandler).Methods("POST", "GET")

	accountRouter.HandleFunc(signupEndpoint, h.SignupHandler).Methods("POST")
	log.Infof("To register a user send a POST request to \"<domain-name>/api/account%s\":\n", signupEndpoint)
	log.Infoln(`{"name":"<new-login>","password": "<new-password>"}`)

	accountRouter.HandleFunc("/{name}", h.GetAccountByNameHandler).Methods("GET")
	accountRouter.HandleFunc("/", h.GetAccountHandler).Methods("GET")
	accountRouter.HandleFunc("/", h.UpdateAccountHandler).Methods("PUT")
	accountRouter.HandleFunc("/", h.DeleteAccountHandler).Methods("DELETE")

	postsRouter := api.PathPrefix("/posts").Subrouter()
	postsRouter.HandleFunc("/info", h.GetPostsInfoHandler).Methods("GET")
	postsRouter.HandleFunc("/{pageNum:[0-9]+}", h.GetPostsHandler).Methods("GET")

	postRouter := api.PathPrefix("/post").Subrouter()
	postRouter.HandleFunc("/{id}", h.GetPostHandler).Methods("GET")
	postRouter.HandleFunc("/", h.CreatePostHandler).Methods("POST")
	postRouter.HandleFunc("/", h.UpdatePostHandler).Methods("PUT")
	postRouter.HandleFunc("/{id}", h.DeletePostHandler).Methods("DELETE")

	// Running the server with a given configuration
	server := &http.Server{
		Handler:      h.CORSMiddleware(conf.Server.OriginAllowed + ", " + conf.Server.Domain)(r), // Setting up CORS middleware
		Addr:         ":" + conf.Server.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, conf.Server.Domain, http.StatusSeeOther)
	})

	log.Fatal(server.ListenAndServe())
}

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
