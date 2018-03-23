package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/lordofthejars/testhub/hub"
)

func FindBuildSummary(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	project := params["project"]
	build := params["build"]

	buildDetail, error := hub.FindBuildDetail(resolveStorageDirectory(), project, build)

	if error != nil {
		sendError(w, error)
		return
	}

	hub.Debug("Finding Summary for project: %s build: %s", project, build)

	json.NewEncoder(w).Encode(buildDetail)
}

func RegisterTestRun(w http.ResponseWriter, r *http.Request) {
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

	testFiles, error := hub.GetTestFiles(fullPath)

	if error != nil {
		sendError(w, error)
		return
	}

	testSuiteResult, error := hub.CreateTestSuite(testFiles)

	if error != nil {
		sendError(w, error)
		return
	}

	error = testSuiteResult.WriteToJson(fullPath)

	w.WriteHeader(http.StatusCreated)

}

func FindBuildsWithStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	currentProject := params["project"]

	hub.Debug("Finding builds for project: %s", currentProject)

	project, error := hub.FindBuildsWithStatus(resolveStorageDirectory(), currentProject)

	if error != nil {
		sendError(w, error)
		return
	}

	json.NewEncoder(w).Encode(project)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) error {
	templates, _ := template.ParseFiles(tmpl)
	return templates.Execute(w, p)
}

func ShowBuildDetailPage(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	project := params["project"]
	build := params["build"]

	hub.Debug("Finding Summary for project: %s build: %s", project, build)

	buildDetail, error := hub.FindBuildDetail(resolveStorageDirectory(), project, build)

	if error != nil {
		sendError(w, error)
		return
	}

	error = renderTemplate(w, "tmpl/details.html", buildDetail)

	if error != nil {
		sendError(w, error)
		return
	}

}

func ShowBuildsPage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	currentProject := params["project"]

	hub.Debug("Finding builds for project: %s", currentProject)

	project, error := hub.FindBuildsWithStatus(resolveStorageDirectory(), currentProject)

	if error != nil {
		sendError(w, error)
		return
	}

	error = renderTemplate(w, "tmpl/builds.html", project)

	if error != nil {
		sendError(w, error)
		return
	}
}

func sendError(w http.ResponseWriter, err error) {
	// Should we add error message in header or somewhere in response?
	hub.Error(err.Error())
	switch err.(type) {
	case *hub.InvalidLocation:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error occured and request couldn't processed.")
	}
}

func resolveStorageDirectory() string {
	usr, _ := user.Current()
	dir := usr.HomeDir

	return filepath.Join(dir, ".hub")
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tmpl/favicon.ico")
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/{project}/{build}", RegisterTestRun).
		Methods("POST").
		Headers("Content-Type", "application/gzip")

	router.HandleFunc("/api/{project}/{build}", FindBuildSummary).
		Methods("GET")

	router.HandleFunc("/api/{project}", FindBuildsWithStatus).
		Methods("GET")

	router.HandleFunc("/{project}", ShowBuildsPage).
		Methods("GET")

	router.HandleFunc("/{project}/{build}", ShowBuildDetailPage).
		Methods("GET")

	router.HandleFunc("/favicon.ico", faviconHandler)
	hub.Info("TestHub Up and Running at %d", 8000)

	http.ListenAndServe(":8000", router)
}
