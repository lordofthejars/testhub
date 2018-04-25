package hub

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yargevad/filepathx"
)

type InvalidLocation struct {
	Location string
}

func (e *InvalidLocation) Error() string {
	return fmt.Sprintf("%s path does not exists", e.Location)
}

type Project struct {
	Project string  `json:"project"`
	Builds  []Build `json:"builds"`
}

func (p Project) orderBuild() {
	sort.Slice(p.Builds, func(i, j int) bool {
		return p.Builds[i].Created.Before(p.Builds[j].Created)
	})
}

func (p Project) GetLastBuild() Build {
	return p.Builds[len(p.Builds)-1]
}

func (p Project) TestTrend() []string {
	var values []string

	for _, build := range p.Builds {
		values = append(values, strconv.Itoa(build.NumberOfTests))
	}
	return values

}

func (p Project) TimeTrend() []string {
	var values []string

	for _, build := range p.Builds {
		values = append(values, strconv.FormatFloat(build.Time, 'E', -1, 64))
	}
	return values
}

func (p Project) TimeTrendJS() template.JS {
	return template.JS(strings.Join(p.TimeTrend(), ","))
}

func (p Project) TestTrendJS() template.JS {
	return template.JS(strings.Join(p.TestTrend(), ","))
}

type Build struct {
	ID            string    `json:"id"`
	Success       bool      `json:"success"`
	NumberOfTests int       `json:"numberTests"`
	Created       time.Time `json:"created"`
	Time          float64   `json:"time"`
	BuildURL      string    `json:"buildUrl"`
}

func (b Build) IsBuildUrlSet() bool {
	return len(b.BuildURL) > 0
}

func (b Build) CreatedTime() string {
	const layout = "Jan 2, 2006 at 3:04pm"
	return b.Created.Format(layout)
}

type BuildDetails struct {
	Build
	Project          string       `json:"project"`
	NumberOfFailures int          `json:"numberFailures"`
	NumberOfErrors   int          `json:"numberErrors"`
	NumberOfSkips    int          `json:"numberSkips"`
	Commit           string       `json:"commit"`
	CommitURL        string       `json:"commitUrl"`
	Branch           string       `json:"branch"`
	BranchURL        string       `json:"branchUrl"`
	RepoURL          string       `json:"repoUrl"`
	Tests            []TestResult `json:"tests"`
}

func (bd BuildDetails) IsCommitSet() bool {
	return len(bd.Commit) > 0
}

func (bd BuildDetails) IsBranchSet() bool {
	return len(bd.Branch) > 0
}

func FindBuildDetail(home string, project string, module string) (BuildDetails, error) {

	buildLocation, err := GetBuildLayout(home, project, module)

	if err != nil {
		return BuildDetails{}, err
	}

	tsr := TestSuiteResult{}
	tsr.LoadFromJson(buildLocation)

	const layout = "Jan 2, 2006 at 3:04pm"
	stat, err := os.Stat(buildLocation)

	if err != nil {
		return BuildDetails{}, nil
	}

	testFiles, err := GetTestFiles(buildLocation)

	if err != nil {
		return BuildDetails{}, nil
	}

	var tests []TestResult

	for _, file := range testFiles {
		var testResult TestResult
		var err error

		switch tsr.Type {
		case SUREFIRE:
			testResult, err = LoadTestResultFromSurefire(file)
		case GRADLE:
			testResult, err = LoadTestResultFromGradle(file)
		}

		if err != nil {
			return BuildDetails{}, err
		}

		tests = append(tests, testResult)
	}

	return BuildDetails{
		Build{module, tsr.IsSuccess(), tsr.Total, stat.ModTime(), tsr.Time, tsr.GetBuildUrl()}, project, tsr.Failures, tsr.Errors, tsr.Skipped, tsr.Commit, tsr.GetCommitUrl(), tsr.Branch, tsr.GetBranchUrl(), tsr.GetRepoUrl(), tests}, nil

}

func FindAllProjects(home string) ([]string, error) {

	var projectNames []string

	if !exists(home) {
		return projectNames, nil
	}

	files, err := ioutil.ReadDir(home)

	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.Mode().IsDir() {
			projectNames = append(projectNames, file.Name())
		}
	}

	return projectNames, nil
}

func FindBuildsWithStatus(home string, project string) (Project, error) {

	buildsLocation, err := GetListOfBuilds(home, project)

	if err != nil {
		return Project{}, err
	}

	var builds []Build

	for _, buildPath := range buildsLocation {
		tsr := TestSuiteResult{}
		moduleLocation := buildPath.Name()
		err := tsr.LoadFromJson(filepath.Join(home, project, moduleLocation))

		if err != nil {
			return Project{}, err
		}

		buildID := moduleLocation[strings.LastIndex(moduleLocation, string(os.PathSeparator))+1:]
		builds = append(builds, Build{buildID, tsr.IsSuccess(), tsr.Total, buildPath.ModTime(), tsr.Time, tsr.BuildUrl})
	}

	projectData := Project{project, builds}
	projectData.orderBuild()
	return projectData, nil

}

func GetTestFiles(destination string) ([]string, error) {

	if exists(destination) {
		pathToExplore := destination + "/**/*.xml"
		return filepathx.Glob(pathToExplore)
	}

	return nil, &InvalidLocation{destination}

}

func GetBuildLayout(home string, project string, build string) (string, error) {

	fullPath := filepath.Join(home, project, build)

	if exists(fullPath) {
		return fullPath, nil
	}

	return "", &InvalidLocation{fullPath}
}

func CreateBuildLayout(home string, project string, build string) (string, error) {
	fullPath := filepath.Join(home, project, build)

	if exists(fullPath) {
		return "", &AlreadyCreatedBuildError{project, build}
	}

	err := os.MkdirAll(fullPath, 0755)

	Debug("Directory %s created to store test results", fullPath)

	return fullPath, err
}

func DeleteBuild(home, project, build string) error {
	fullPath := filepath.Join(home, project, build)
	if exists(fullPath) {
		os.RemoveAll(fullPath)
		return nil
	}

	return &InvalidLocation{fullPath}
}

func GetListOfBuilds(home string, project string) ([]os.FileInfo, error) {
	fullPath := filepath.Join(home, project)
	if exists(fullPath) {
		return ioutil.ReadDir(fullPath)
	}

	return nil, &InvalidLocation{fullPath}
}

func exists(path string) bool {

	if _, err := os.Stat(path); err == nil {
		return true
	}

	// Maybe does not exists or there are any permission problem but in anyeway for us is that it does not exists
	return false
}

func UncompressContent(destination string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	defer gzr.Close()
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			Debug("Test Results uncompressed at %s", destination)
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(destination, header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()
			Debug("Uncompressing %s", target)
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
}
