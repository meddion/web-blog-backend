package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"log"
)

type response struct {
	Ok   bool        `json:"ok"`
	Err  string      `json:"error"`
	Body interface{} `json:"body"`
}

func sendErrorResp(w http.ResponseWriter, err string, code int) {
	if code == http.StatusInternalServerError && os.Getenv("DEV_STAGE") == "" {
		log.Println(err)
		err = "on getting an internal server error"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response{
		Err: err,
	})
}

func sendSuccessResp(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response{
		Ok:   true,
		Body: body,
	})
}
