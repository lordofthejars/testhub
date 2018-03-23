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

type TestResult struct {
	Name    string
	Success bool
	Total,
	Failures,
	Errors,
	Skipped int
}

func (tsr TestResult) IsSuccess() bool {
	return tsr.Failures == 0 && tsr.Errors == 0
}

func (tsr *TestSuiteResult) countTestResult(testResult *testresultparser.TestResults) {
	tsr.Total += testResult.Summary.Total
	tsr.Failures += testResult.Summary.Failures
	tsr.Errors += testResult.Summary.Errors
	tsr.Skipped += testResult.Summary.Skipped
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

func LoadTestResult(filepath string) (TestResult, error) {

	tr, err := parseSurefireReport(filepath)

	if err != nil {
		return TestResult{}, err
	}

	return TestResult{tr.Name, tr.Summary.Failures+tr.Summary.Errors == 0, tr.Summary.Total, tr.Summary.Failures, tr.Summary.Errors, tr.Summary.Skipped}, nil
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
