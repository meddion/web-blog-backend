package main

import (
	"math/rand"
	"net/http"
	"time"

	"log"

	"github.com/gorilla/mux"
	"github.com/meddion/web-blog/pkg/config"
	h "github.com/meddion/web-blog/pkg/handlers"
	"github.com/meddion/web-blog/pkg/models"
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

	// Redirecting to client
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, conf.Server.Domain, http.StatusSeeOther)
	})

	// Setting up endpoints with /api prefix in common
	api := r.PathPrefix("/api").Subrouter()
	// Serving static files and manipulating with them
	s := h.NewStaticHandler(conf.Server.StaticDir)
	r.HandleFunc("/static/{path:.*}", s.StaticHandler).Methods("GET")
	static := api.PathPrefix("/static").Subrouter()
	static.HandleFunc("/{path:.*}", s.AddFileHandler).Methods("POST")
	static.HandleFunc("/{path:.*}", s.DeleteFileHandler).Methods("DELETE")
	static.HandleFunc("/filenames/{path:.*}", s.GetFilenamesHandler).Methods("GET")

	accountRouter := api.PathPrefix("/account").Subrouter()
	accountRouter.HandleFunc("/login", h.LoginHandler).Methods("POST")
	accountRouter.HandleFunc("/logout", h.LogoutHandler).Methods("POST", "GET")

	signupHash := genRandSeqOfLen(32)
	accountRouter.HandleFunc("/signup/"+signupHash, h.SignupHandler).Methods("POST")
	log.Printf("To register follow \"/api/account/signup/%s\"", signupHash)
	log.Printf(`{"name":"<new-login>","password": "<new-password>"}`)

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

	// Setting up our session-auth middleware
	// Passing routes that do not require authorization to NewSessionAuthMiddleware
	sessionAuthMiddleware, err := h.NewSessionAuthMiddleware(
		"/static/{path:.*}",
		"/api/static/filenames/{path:.*}",
		"/api/account/login",
		"/api/account/signup/"+signupHash,
		"/api/account/{name}",
		"/api/posts/info",
		"/api/posts/{pageNum:[0-9]+}",
		"/api/post/{id}",
	)
	if err != nil {
		log.Panic(err)
	}
	r.Use(sessionAuthMiddleware.Middleware)

	// Running the server with a given configuration
	server := &http.Server{
		Handler:      h.CORSMiddleware(conf.Server.OriginAllowed)(r), // Setting up CORS middleware
		Addr:         ":" + conf.Server.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

func genRandSeqOfLen(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
