package handlers

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	"github.com/meddion/web-blog/pkg/models"
)

// StaticHandler serves static files from [staticPath] folder
func StaticHandler(w http.ResponseWriter, r *http.Request) {
	dir, filename, ext := extractDirFilenameExt(mux.Vars(r)["path"])
	file := models.File{Dir: dir, Name: filename, Ext: ext}
	// Return the file from DB
	if err := file.Get(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", mime.TypeByExtension("."+file.Ext))
	if _, err := w.Write(file.File.Data); err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteFileHandler is used for deleting files from the static dir
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	if isPathEmpty(mux.Vars(r)["path"]) {
		sendErrorResp(w, "the path to the file is empty", http.StatusBadRequest)
		return
	}
	dir, filename, ext := extractDirFilenameExt(mux.Vars(r)["path"])
	file := models.NewEmptyFile(dir, filename, ext)
	// Remove file from DB
	if res, err := file.Delete(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
	} else if res.DeletedCount == 0 {
		sendErrorResp(w, "the file wasn't found, thus wasn't deleted", http.StatusBadRequest)
	} else {
		sendSuccessResp(w, nil)
	}
}

// AddFileHandler is used for adding files to the static dir.
// If the file with a specified name and extension
// already exists in the folder - replace it.
func AddFileHandler(w http.ResponseWriter, r *http.Request) {
	// Getting the byte slice from the request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Extracting directory, filename and extension info
	dir, filename, ext := extractDirFilenameExt(mux.Vars(r)["path"])
	if filename == "" {
		filename = fmt.Sprintf("%d", time.Now().Unix())
	}
	if ext == "" {
		ext = strings.TrimLeft(mimetype.Detect(data).Extension(), ".")
	}

	file := models.NewFile(dir, filename, ext, data)
	// Saving the file as a blob into DB
	if _, err = file.Save(r.Context()); err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccessResp(w, nil)
}

// GetFilenamesHandler is used for getting filenames in folders with a specific extention
func GetFilenamesHandler(w http.ResponseWriter, r *http.Request) {
	dir, ext, _ := extractDirFilenameExt(mux.Vars(r)["path"])

	filenames, err := models.ListFilenamesWhere(r.Context(), dir, ext)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sendSuccessResp(w, map[string][]string{
		"filenames": filenames,
	})
}

func isPathEmpty(path string) bool {
	return filepath.Clean(path) == "."
}

func extractDirFilenameExt(rawPath string) (string, string, string) {
	path := filepath.Clean(rawPath)
	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	if filename == dir {
		filename = ""
	}
	if dir == "." {
		dir = "/"
	}
	var extension string
	if strings.Contains(filename, ".") {
		if part := strings.TrimLeft(filename, "."); part != filename {
			filename = ""
			extension = part
		} else {
			parts := strings.Split(filename, ".")
			filename = parts[0]
			if len(parts) > 1 {
				extension = parts[1]
			}
		}
	}

	return dir, filename, extension
}
