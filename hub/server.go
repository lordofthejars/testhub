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
	"github.com/lordofthejars/testhub/auth"
)

type AlreadyCreatedBuildError struct {
	Project string
	Build   string
}

func (e *AlreadyCreatedBuildError) Error() string {
	return fmt.Sprintf("Project %s Build %s has been already published results", e.Project, e.Build)
}

type AuthenticationError struct {
	Username string
	Message  string
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("Error authenticating %s user with %s", e.Username, e.Message)
}

var users *auth.Users
var securityEnabled = false
var configuration *Config

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

func login(w http.ResponseWriter, r *http.Request) {
	Debug("Login to server")

	var content interface{}
	json.NewDecoder(r.Body).Decode(&content)

	contentAsMap := content.(map[string]interface{})

	username, ok := contentAsMap["username"].(string)

	if !ok {
		sendError(w, &AuthenticationError{username, "Username field not found"})
	}

	password, ok := contentAsMap["password"].(string)
	if !ok {
		sendError(w, &AuthenticationError{username, "Password field not found"})
	}

	if users != nil {
		if users.ValidateUser(username, password) {
			token, err := auth.GenerateToken(username, "")

			if err != nil {
				sendError(w, err)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, token)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
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

func StartServer(config *Config) {

	initializeUsers(*config)
	configuration = config

	router := mux.NewRouter()

	router.HandleFunc("/api/login", login).
		Methods("POST")

	router.HandleFunc("/api/project/{project}/{build}", auth.WithJWT(configuration.Authentication.Secret, securityEnabled, registerSurefireTestRun)).
		Methods("POST").
		Headers("Content-Type", "application/gzip", "x-testhub-type", "surefire")

	router.HandleFunc("/api/project/{project}/{build}", findBuildSummary).
		Methods("GET")

	router.HandleFunc("/api/project/{project}/{build}", auth.WithJWT(configuration.Authentication.Secret, securityEnabled, deleteBuild)).
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
	Info("TestHub Up and Running at %d and repository %s and security enabled %t", configuration.Port, configuration.Repository.Path, securityEnabled)

	if configuration.isSSLConfigured() {
		http.ListenAndServeTLS(":"+strconv.Itoa(configuration.Port), configuration.Cert, configuration.Key, router)
	} else {
		http.ListenAndServe(":"+strconv.Itoa(configuration.Port), router)
	}
}

func initializeUsers(configuration Config) {
	users = auth.ReadUsersFromFile(configuration.Authentication.UsersPath)

	if users.AreUsers() {
		securityEnabled = true
	}

}
