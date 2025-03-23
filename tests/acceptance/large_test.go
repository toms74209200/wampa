//go:build large
// +build large

package acceptance

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestLargeFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths: []string{
				"../../features/remote_file_handling.feature",
			},
			Tags:     "@large",
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run large feature tests")
	}
}
