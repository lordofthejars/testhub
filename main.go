package main

import (
	"fmt"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/lordofthejars/testhub/hub"
)

func UnCompressTestRun(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	project := params["project"]
	build := params["build"]

	fullPath, error := hub.CreateBuildLayout(resolveStorageDirectory(), project, build)

	if error != nil {
		sendError(w, error)
		return
	}

	error = hub.UncompressContent(fullPath, r.Body)

	if error != nil {
		sendError(w, error)
		return
	}
}

func sendError(w http.ResponseWriter, err error) {
	// Should we add error message in header or somewhere in response?
	hub.Error(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Error occured and request couldn't processed.")
}

func resolveStorageDirectory() string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	return filepath.Join(dir, ".hub")
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{project}/{build}", UnCompressTestRun).
		Methods("POST").
		Headers("Content-Type", "application/gzip")

	hub.Info("TestHub Up and Running at %d", 8000)

	http.ListenAndServe(":8000", router)
}
