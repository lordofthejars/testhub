package hub

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func findBuildSummary(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	project := params["project"]
	build := params["build"]

	buildDetail, error := FindBuildDetail(resolveStorageDirectory(), project, build)

	if error != nil {
		sendError(w, error)
		return
	}

	Debug("Finding Summary for project: %s build: %s", project, build)

	json.NewEncoder(w).Encode(buildDetail)
}

func registerSurefireTestRun(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	project := params["project"]
	build := params["build"]

	fullPath, error := CreateBuildLayout(resolveStorageDirectory(), project, build)

	if error != nil {
		sendError(w, error)
		return
	}

	error = UncompressContent(fullPath, r.Body)

	if error != nil {
		sendError(w, error)
		return
	}

	testFiles, error := GetTestFiles(fullPath)

	if error != nil {
		sendError(w, error)
		return
	}

	testSuiteResult, error := CreateTestSuite(testFiles)

	if error != nil {
		sendError(w, error)
		return
	}

	error = testSuiteResult.WriteToJson(fullPath)

	w.WriteHeader(http.StatusCreated)

}

func deleteBuild(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	project := params["project"]
	build := params["build"]

	Debug("Deleting build %s of project: %s", build, project)
	error := DeleteBuild(resolveStorageDirectory(), project, build)

	if error != nil {
		sendError(w, error)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func findBuildsWithStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	currentProject := params["project"]

	Debug("Finding builds for project: %s", currentProject)

	project, error := FindBuildsWithStatus(resolveStorageDirectory(), currentProject)

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

func showBuildDetailPage(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	project := params["project"]
	build := params["build"]

	Debug("Finding Summary for project: %s build: %s", project, build)

	buildDetail, error := FindBuildDetail(resolveStorageDirectory(), project, build)

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

func showBuildsPage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	currentProject := params["project"]

	Debug("Finding builds for project: %s", currentProject)

	project, error := FindBuildsWithStatus(resolveStorageDirectory(), currentProject)

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
	Error(err.Error())
	switch err.(type) {
	case *InvalidLocation:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error occured and request couldn't processed.")
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tmpl/favicon.ico")
}

func StartServer(configuration *Config) {

	router := mux.NewRouter()

	router.HandleFunc("/api/{project}/{build}", registerSurefireTestRun).
		Methods("POST").
		Headers("Content-Type", "application/gzip", "x-testhub-type", "surefire")

	router.HandleFunc("/api/{project}/{build}", findBuildSummary).
		Methods("GET")

	router.HandleFunc("/api/{project}/{build}", deleteBuild).
		Methods("DELETE")

	router.HandleFunc("/api/{project}", findBuildsWithStatus).
		Methods("GET")

	router.HandleFunc("/{project}", showBuildsPage).
		Methods("GET")

	router.HandleFunc("/{project}/{build}", showBuildDetailPage).
		Methods("GET")

	router.HandleFunc("/favicon.ico", faviconHandler)
	Info("TestHub Up and Running at %d and repository %s", configuration.Port, configuration.Repository.Path)

	http.ListenAndServe(":"+strconv.Itoa(configuration.Port), router)
}
