package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
)

// StaticHandler serves static files
type StaticHandler struct {
	staticPath string
}

// NewStaticHandler is constructor for StaticHandler
func NewStaticHandler(staticPath string) StaticHandler {
	staticPath, err := filepath.Abs(staticPath)
	if err != nil {
		panic(err)
	}
	return StaticHandler{staticPath}
}

func (s StaticHandler) isPathEmpty(path string) bool {
	return filepath.Clean(path) == ""
}

func (s StaticHandler) getFullPath(rawPath string) string {
	return filepath.Join(s.staticPath, rawPath)
}

// StaticHandler serves static files from [staticPath] folder
func (s StaticHandler) StaticHandler(w http.ResponseWriter, r *http.Request) {
	path := s.getFullPath(mux.Vars(r)["path"])
	// check whether a file exists at the given path
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist
		sendErrorResp(w, "on finding the file", http.StatusNotFound)
		return
	} else if err != nil {
		sendErrorResp(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, path)

}

// DeleteFileHandler is used for deleting files from the static dir
func (s StaticHandler) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	if s.isPathEmpty(mux.Vars(r)["path"]) {
		sendErrorResp(w, "on deleting a container folder", http.StatusBadRequest)
		return
	}
	path := s.getFullPath(mux.Vars(r)["path"])
	if err := os.Remove(path); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendSuccessResp(w, nil)
}

// AddFileHandler is used for adding files to the static dir.
// If the file with a specified name and extension
// already exists in the folder - replace it.
func (s StaticHandler) AddFileHandler(w http.ResponseWriter, r *http.Request) {
	path := s.getFullPath(mux.Vars(r)["path"])
	// Get byte slice from the request body
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Detecting filename and extension
	var filename, extension string
	structure := strings.Split(path, "/")
	if fparts := strings.Split(structure[len(structure)-1], "."); len(fparts) > 1 {
		filename = fparts[0]
		if fparts[1] != "" {
			extension = "." + fparts[1]
		}
		path = strings.Join(structure[:len(structure)-1], "/")
	} else {
		path = strings.Join(structure, "/")
	}

	if extension == "" {
		extension = mimetype.Detect(data).Extension()
	}
	if filename == "" {
		filename = fmt.Sprintf("%d", time.Now().Unix())
	}
	filename = filename + extension

	// Creating folder structure if needed
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Writing the request body to a new created file
	if err = ioutil.WriteFile(filepath.Join(path, filename), data, os.ModePerm); err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	sendSuccessResp(w, nil)
}

// GetFilenamesHandler is used for getting filenames in folders with a specific extention
func (s StaticHandler) GetFilenamesHandler(w http.ResponseWriter, r *http.Request) {
	if s.isPathEmpty(mux.Vars(r)["path"]) {
		sendErrorResp(w, "on getting an empty path value", http.StatusBadRequest)
		return
	}
	path := s.getFullPath(mux.Vars(r)["path"])
	slice := strings.Split(path, "/")
	// Getting the extention from the path paremeter
	extension := slice[len(slice)-1]
	// The rest is the folder path we wanna look at
	dir := strings.Join(slice[:len(slice)-1], "/")

	filenames, err := ioutil.ReadDir(dir)
	if err != nil {
		sendErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	fileList := make([]string, 0)
	for _, f := range filenames {
		if f.IsDir() {
			continue
		}
		filename := f.Name()
		if extension != "*" && !strings.HasSuffix(filename, extension) {
			continue
		}
		fileList = append(fileList, filename)
	}
	sendSuccessResp(w, map[string][]string{
		"filenames": fileList,
	})
}
