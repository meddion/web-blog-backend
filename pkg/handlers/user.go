package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meddion/web-blog/pkg/models"
	"github.com/meddion/web-blog/pkg/session"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handlers which do not require user to be authorized

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Checking if a user is already logged in

	session, ok := r.Context().Value("session").(session.Session)
	if !ok {
		sendErrorResp(w, "on retrieving a session from a request's context", http.StatusInternalServerError)
		return
	}
	if session.IsValuePresent("USER") {
		sendSuccessResp(w, nil)
		return
	}

	user := &models.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		sendErrorResp(w, "on decoding a request body", http.StatusBadRequest)
		return
	}
	if err := user.ValidateLoginForm(); err != nil {
		sendErrorResp(w, "on matching the credentials for a user", http.StatusUnauthorized)
		return
	}
	passwordFromRequest := user.Password
	if err := user.Get(r.Context()); err != nil {
		if err == mongo.ErrNoDocuments {
			sendErrorResp(w, "on matching the credentials for a user", http.StatusUnauthorized)
			return
		}
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	match, err := models.CompareHashAndPassword(user.Password, passwordFromRequest)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !match {
		sendErrorResp(w, "on matching the credentials for a user", http.StatusUnauthorized)
		return
	}
	// Creates a session & puts user's object there
	if err := session.Set("USER", user); err != nil {
		sendErrorResp(w, "on saving a user's object into session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccessResp(w, nil)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Checking if a user is already logged in
	session, ok := r.Context().Value("session").(session.Session)
	if !ok {
		sendErrorResp(w, "on retrieving a session from a request's context", http.StatusInternalServerError)
		return
	}
	if session.IsValuePresent("USER") {
		sendErrorResp(w, "", http.StatusBadRequest)
		return
	}

	user := &models.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		sendErrorResp(w, "on decoding a request body", http.StatusBadRequest)
		return
	}
	if err := user.ValidateSignupForm(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := user.Create(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Creates a session & puts user's object there
	if err := session.Set("USER", user); err != nil {
		sendErrorResp(w, "on saving a user's object into session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sendSuccessResp(w, nil)
}

func GetAccountByNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user, err := models.GetUserInfoByName(r.Context(), vars["name"])
	if err != nil {
		sendErrorResp(w, "on founding the user", http.StatusNotFound)
		return
	}
	sendSuccessResp(w, user)
}

// Require authorization

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Checking if a *session.Manager instance was passed
	manager, ok := r.Context().Value("manager").(*session.Manager)
	if !ok {
		sendErrorResp(w, "on retrieving a manager (*session.Manager) from a request's context", http.StatusInternalServerError)
		return
	}
	manager.SessionDestroy(w, r)
	sendSuccessResp(w, nil)
}

func GetAccountHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	copyUser := *user
	copyUser.Password = ""
	sendSuccessResp(w, copyUser)
}

func UpdateAccountHandler(w http.ResponseWriter, r *http.Request) {
	session, err := GetSession(r)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, ok := session.Get("USER").(*models.User)
	if !ok {
		sendErrorResp(w, "on founding a user's object in the session", http.StatusInternalServerError)
		return
	}

	newUser := &models.User{}
	if err := json.NewDecoder(r.Body).Decode(newUser); err != nil {
		sendErrorResp(w, "on decoding a request body", http.StatusBadRequest)
		return
	}
	if err := newUser.ValidateUpdateForm(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	newUser.ID = user.ID
	if err := newUser.Update(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Putting new user's data back to the session
	user.Assign(newUser)
	if err := session.Set("USER", user); err != nil {
		sendErrorResp(w, "on saving a user's object into session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccessResp(w, nil)
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserFromSession(r)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := models.DeleteUserByID(r.Context(), user.ID); err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/api/account/logout", http.StatusSeeOther)
}

func GetSession(r *http.Request) (session.Session, error) {
	session, ok := r.Context().Value("session").(session.Session)
	if !ok {
		return nil, errors.New("on retrieving a session from a request's context")
	}
	return session, nil
}

func GetUserFromSession(r *http.Request) (*models.User, error) {
	session, err := GetSession(r)
	if err != nil {
		return nil, err
	}
	user, ok := session.Get("USER").(*models.User)
	if !ok {
		return nil, errors.New("on founding a user's object in the session")
	}
	return user, nil
}
