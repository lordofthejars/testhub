package hub

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
)

type AlreadyCreatedBuild struct {
	Project string
	Build   string
}

func (e *AlreadyCreatedBuild) Error() string {
	return fmt.Sprintf("Project %s Build %s has been already published results", e.Project, e.Build)
}

var box = packr.NewBox("../tmpl")

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

	if exists(fullPath) {
		sendError(w, &AlreadyCreatedBuild{project, build})
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

	repo := getSingleQueryParam(r.URL.Query(), "repoUrl")
	branch := getSingleQueryParam(r.URL.Query(), "branch")
	commit := getSingleQueryParam(r.URL.Query(), "commit")
	buildLocation := getSingleQueryParam(r.URL.Query(), "buildUrl")
	repoType := getSingleQueryParam(r.URL.Query(), "repoType")

	testSuiteResult.Branch = branch
	testSuiteResult.RepoUrl = repo
	testSuiteResult.Commit = commit
	testSuiteResult.BuildUrl = buildLocation
	testSuiteResult.RepoType = repoType

	error = testSuiteResult.WriteToJson(fullPath)

	w.WriteHeader(http.StatusCreated)

}

func getSingleQueryParam(queryParams url.Values, name string) string {
	keys, err := queryParams[name]

	if !err || len(keys) < 1 {
		return ""
	}

	return keys[0]
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

func renderTemplate(w http.ResponseWriter, tmplName string, p interface{}) error {
	html, err := box.MustString(tmplName)

	if err != nil {
		return err
	}

	templates, _ := template.New(tmplName).Parse(html)
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

	error = renderTemplate(w, "details.html", buildDetail)

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

	error = renderTemplate(w, "builds.html", project)

	if error != nil {
		sendError(w, error)
		return
	}
}

func showProjects(w http.ResponseWriter, r *http.Request) {

	Debug("Finding all projects")

	projects, error := FindAllProjects(resolveStorageDirectory())

	if error != nil {
		sendError(w, error)
		return
	}

	error = renderTemplate(w, "projects.html", projects)

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
		fmt.Fprintf(w, "Error occured and request couldn't be processed: %s.", err.Error())
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "tmpl/favicon.ico")
}

func StartServer(configuration *Config) {

	router := mux.NewRouter()

	router.HandleFunc("/api/project/{project}/{build}", registerSurefireTestRun).
		Methods("POST").
		Headers("Content-Type", "application/gzip", "x-testhub-type", "surefire")

	router.HandleFunc("/api/project/{project}/{build}", findBuildSummary).
		Methods("GET")

	router.HandleFunc("/api/project/{project}/{build}", deleteBuild).
		Methods("DELETE")

	router.HandleFunc("/api/project/{project}", findBuildsWithStatus).
		Methods("GET")

	router.HandleFunc("/", showProjects).
		Methods("GET")

	router.HandleFunc("/project", showProjects).
		Methods("GET")

	router.HandleFunc("/project/{project}", showBuildsPage).
		Methods("GET")

	router.HandleFunc("/project/{project}/{build}", showBuildDetailPage).
		Methods("GET")

	router.HandleFunc("/favicon.ico", faviconHandler)
	Info("TestHub Up and Running at %d and repository %s", configuration.Port, configuration.Repository.Path)

	if configuration.isSSLConfigured() {
		http.ListenAndServeTLS(":"+strconv.Itoa(configuration.Port), configuration.Cert, configuration.Key, router)
	} else {
		http.ListenAndServe(":"+strconv.Itoa(configuration.Port), router)
	}
}
