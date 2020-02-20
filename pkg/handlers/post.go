package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/meddion/web-blog/pkg/models"
)

const postsPerPage int64 = 10

// Handlers which do not require user to be authorized

func GetPostsInfoHandler(w http.ResponseWriter, r *http.Request) {
	totalNumOfPosts, err := models.GetTotalNumOfPosts(r.Context(), nil)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccessResp(w, map[string]int64{
		"totalNumOfPosts": totalNumOfPosts,
		"postsPerPage":    postsPerPage,
	})
}

func GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageNum, err := strconv.ParseInt(vars["pageNum"], 10, 64)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	if pageNum < 1 {
		pageNum = 1
	}
	posts, err := models.GetPosts(
		r.Context(),
		models.CreatePostsPipeline(r, postsPerPage, pageNum),
	)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	if posts == nil || len(posts) < 1 {
		sendErrorResp(w, "nothing was found", http.StatusBadRequest)
		return
	}
	sendSuccessResp(w, posts)
}

func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	post, err := models.GetPostByID(r.Context(), vars["id"])
	if err != nil {
		sendErrorResp(w, "on not getting any posts from db", http.StatusBadRequest)
		return
	}
	sendSuccessResp(w, post)
}

// Require authorization

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := &models.Post{}
	if err := json.NewDecoder(r.Body).Decode(post); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := GetUserFromSession(r)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post.AuthorID = user.ID
	if err := post.Save(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendSuccessResp(w, map[string]interface{}{"id": post.ID})
}

func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := &models.Post{}
	if err := json.NewDecoder(r.Body).Decode(post); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := post.Update(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendSuccessResp(w, nil)
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := models.DeletePostById(r.Context(), vars["id"]); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendSuccessResp(w, nil)
}
