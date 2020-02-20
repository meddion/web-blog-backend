package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/meddion/web-blog/pkg/session"
	_ "github.com/meddion/web-blog/pkg/session/providers/memory"
)

// CORSMiddleware provides Cross-Origin Resource Sharing middleware.
// based on gorilla/handlers implementation
func CORSMiddleware(originAllowed string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Connection", "keep-alive")
			w.Header().Add("Access-Control-Allow-Origin", originAllowed)
			w.Header().Add("Access-Control-Allow-Credentials", "true")
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Accept-Encoding")
			if r.Method == "OPTIONS" {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type sessionAuthMiddleware struct {
	notAuth map[string]struct{}
	manager *session.Manager
}

func NewSessionAuthMiddleware(notAuthURLs ...string) (*sessionAuthMiddleware, error) {
	m := &sessionAuthMiddleware{}
	m.notAuth = make(map[string]struct{})
	for _, val := range notAuthURLs {
		m.notAuth[val] = struct{}{}
	}
	var err error
	m.manager, err = session.NewManager("memory", "SESSION_ID", 60)
	if err != nil {
		return nil, fmt.Errorf("on initializing globalSessionManager: %s", err.Error())
	}
	go m.manager.GC()
	return m, nil
}

func (m *sessionAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		if route == nil {
			sendErrorResp(w, "on getting the route right", http.StatusInternalServerError)
			return
		}
		path, err := route.GetPathTemplate()
		if err != nil {
			sendErrorResp(w, "on getting the path template from the route", http.StatusInternalServerError)
			return
		}
		// Creating a new session.
		session, err := m.manager.SessionStart(w, r)
		if err != nil {
			sendErrorResp(w, "on starting a session for a client", http.StatusInternalServerError)
			return
		}
		switch path {
		case "/api/account/logout":
			r = r.WithContext(context.WithValue(r.Context(), "manager", m.manager))
		default:
			r = r.WithContext(context.WithValue(r.Context(), "session", session))
		}
		// Setting timeout for database operations
		ctxWithTimeout, _ := context.WithTimeout(r.Context(), 3*time.Second)
		r = r.WithContext(ctxWithTimeout)

		// Iterating through URI-paths which do not require an authentication from a user
		if _, ok := m.notAuth[path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		// Sending 401 code if a user is not unauthorized
		if !session.IsValuePresent("USER") {
			sendErrorResp(w, "", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
