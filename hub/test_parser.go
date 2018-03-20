package hub

import (
	"github.com/almighty/almighty-test-runner/testresultparser"
)

func ParseSurefireReport(filepath string) (*testresultparser.TestResults, error) {
	surefireParser := testresultparser.SurefireParser{}
	return surefireParser.Parse(filepath)
}
