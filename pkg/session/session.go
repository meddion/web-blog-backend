package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var providers = make(map[string]Provider)

// Session is an interface which defines session in our system
type Session interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Delete(key interface{}) error
	IsValuePresent(ket interface{}) bool
	GetSessionID() string
}

// Provider is an interface which represents the underlying structure for our Session
type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionUpdate(sid string) error
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64) // sessions's expiry garbage collector
}

func Register(name string, provider Provider) {
	if provider == nil {
		log.Panicf("on registering a provider (%s) with a nil-value", name)
	}
	if _, dup := providers[name]; dup {
		log.Panicf("on registering a provider with a duplicate name: %s", name)
	}
	providers[name] = provider
}

// Manager is an API for manipulating with sessions,
// abstracted from the internal implementation of a storage
type Manager struct {
	provider        Provider
	cookieName      string
	sessionLifeTime int64
	sync.Mutex
}

func NewManager(providerName, cookieName string, sessionLifeTime int64) (*Manager, error) {
	provider, ok := providers[providerName]
	if !ok {
		return nil, fmt.Errorf("on getting an unknown provider for s: %q", providerName)
	}
	return &Manager{
		provider:        provider,
		cookieName:      cookieName,
		sessionLifeTime: sessionLifeTime,
	}, nil
}

// SessionStart an entry point for any page that rely on sessions
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (Session, error) {
	manager.Lock()
	defer manager.Unlock()
	// If a session-cookie doesn't exist, then create a new session
	// & save it as a cookie on a client
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		session, err := manager.sessionCreate(w)
		if err != nil {
			return nil, err
		}
		return session, nil
	}
	// Instead, if the session-cookie does exist - just retrieve it from a provider
	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}
	session, err := manager.provider.SessionRead(sid)
	if err != nil {
		session, err = manager.sessionCreate(w)
		if err != nil {
			return nil, err
		}
	}
	return session, nil
}

// sessionCreate returns a newly created session (with an error)
// and setts a cookie on a client containing that session's id
func (manager *Manager) sessionCreate(w http.ResponseWriter) (Session, error) {
	sid, err := manager.createUniqueSessionID()
	if err != nil {
		return nil, err
	}
	session, err := manager.provider.SessionInit(sid)
	if err != nil {
		return nil, err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     manager.cookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: false,
		MaxAge:   0, // a cookie won't last after closing a browser
	})
	return session, nil
}

// SessionDestroy removes a session-cookie on a client (if it's present)
// & purges a session from a provider by its ID (got from the cookie)
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	manager.Lock()
	defer manager.Unlock()
	manager.provider.SessionDestroy(cookie.Value)
	expiration := time.Now()
	http.SetCookie(w, &http.Cookie{
		Name:     manager.cookieName,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiration,
		MaxAge:   -1,
	})
}

// createUniqueSessionID generates a unique string for our sessions
func (manager *Manager) createUniqueSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GC is a garbage collector for our expired sessions
// it invokes SessionGC() method on a provider after manager.cookieLifeTime time
func (manager *Manager) GC() {
	manager.Lock()
	defer manager.Unlock()
	manager.provider.SessionGC(manager.sessionLifeTime)
	time.AfterFunc(60*time.Second, func() { manager.GC() })
}
