package hub

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/almighty/almighty-test-runner/testresultparser"
)

type TestSuiteResult struct {
	Total,
	Failures,
	Errors,
	Skipped int
	Time float64
	Branch,
	Commit,
	RepoUrl,
	RepoType,
	BuildUrl string
}

func (tsr TestSuiteResult) AnyFailure() bool {
	return tsr.Failures > 0
}

func (tsr TestSuiteResult) AnyError() bool {
	return tsr.Errors > 0
}

func (tsr TestSuiteResult) AnySkipped() bool {
	return tsr.Skipped > 0
}

func (tsr TestSuiteResult) IsSuccess() bool {
	return tsr.Failures == 0 && tsr.Errors == 0
}

func (tsr TestSuiteResult) IsBranchSet() bool {
	return len(tsr.Branch) > 0
}

func (tsr TestSuiteResult) IsCommitSet() bool {
	return len(tsr.Commit) > 0
}

func (tsr TestSuiteResult) IsRepoUrlSet() bool {
	return len(tsr.RepoUrl) > 0
}

func (tsr TestSuiteResult) IsBuildUrlSet() bool {
	return len(tsr.BuildUrl) > 0
}

func (tsr TestSuiteResult) GetRepoUrl() string {
	if tsr.IsRepoUrlSet() {
		return tsr.RepoUrl
	}

	return "#"
}

func (tsr TestSuiteResult) GetBuildUrl() string {
	if tsr.IsBuildUrlSet() {
		return tsr.BuildUrl
	}

	return "#"
}

func (tsr TestSuiteResult) GetCommitUrl() string {
	if tsr.IsCommitSet() && tsr.IsBuildUrlSet() {
		return convertToCorrectCommitUrl(tsr.RepoUrl, tsr.Commit, tsr.RepoType)
	}
	return "#"
}
func (tsr TestSuiteResult) GetBranchUrl() string {
	if tsr.IsBranchSet() && tsr.IsRepoUrlSet() {
		return convertToCorrectBranchUrl(tsr.RepoUrl, tsr.Branch, tsr.RepoType)
	}
	return "#"
}

func (tsr *TestSuiteResult) countTestResult(testResult *testresultparser.TestResults) {
	tsr.Total += testResult.Summary.Total
	tsr.Failures += testResult.Summary.Failures
	tsr.Errors += testResult.Summary.Errors
	tsr.Skipped += testResult.Summary.Skipped
	tsr.Time += testResult.Summary.Time
}

func (tsr *TestSuiteResult) LoadFromJson(destination string) error {
	fullPath := filepath.Join(destination, "testsuite.json")

	if exists(fullPath) {
		content, err := ioutil.ReadFile(fullPath)

		if err != nil {
			return err
		}

		err = json.Unmarshal(content, tsr)
		return err
	}

	return &InvalidLocation{fullPath}
}

func (tsr TestSuiteResult) WriteToJson(destination string) error {
	content, err := json.Marshal(&tsr)

	if err != nil {
		return err
	}

	fullPath := filepath.Join(destination, "testsuite.json")
	err = ioutil.WriteFile(fullPath, content, 0644)

	Debug("Summary generated at %s", fullPath)
	return err
}

type TestResult struct {
	Name    string
	Success bool
	Total,
	Failures,
	Errors,
	Skipped int
	TestMethods []TestMethod
}

type Result uint8

const (
	PASSED Result = iota + 1
	SKIPPED
	FAILURE
	ERROR
)

type TestMethod struct {
	TestCase string
	Time     float64
	Return   Result
	Type     string
	Message  string
	Details  string
}

func (tm TestMethod) IsTypeSet() bool {
	return len(tm.Type) > 0
}

func (tm TestMethod) IsPassed() bool {
	return tm.Return == PASSED
}

func (tm TestMethod) IsSkipped() bool {
	return tm.Return == SKIPPED
}

func (tm TestMethod) IsFailure() bool {
	return tm.Return == FAILURE
}

func (tm TestMethod) IsError() bool {
	return tm.Return == ERROR
}

func toResult(kind testresultparser.TestResultKind) Result {
	switch kind {
	case 1:
		return PASSED
	case 2:
		return SKIPPED
	case 3:
		return FAILURE
	case 4:
		return ERROR
	}

	return 0
}

func LoadTestResult(filepath string) (TestResult, error) {

	tr, err := parseSurefireReport(filepath)

	if err != nil {
		return TestResult{}, err
	}

	var testMethods []TestMethod

	for _, result := range tr.Results {
		testMethods = append(testMethods, TestMethod{result.TestCase, result.Time, toResult(result.Kind), result.Type, result.Message, result.Details})
	}

	return TestResult{tr.Name, tr.Summary.Failures+tr.Summary.Errors == 0, tr.Summary.Total, tr.Summary.Failures, tr.Summary.Errors, tr.Summary.Skipped, testMethods}, nil
}

func CreateTestSuite(files []string) (TestSuiteResult, error) {
	tsr := TestSuiteResult{}
	for _, file := range files {
		tr, err := parseSurefireReport(file)

		if err != nil {
			return tsr, err
		}
		tsr.countTestResult(tr)
	}
	return tsr, nil
}

func parseSurefireReport(filepath string) (*testresultparser.TestResults, error) {
	surefireParser := testresultparser.SurefireParser{}
	return surefireParser.Parse(filepath)
}
