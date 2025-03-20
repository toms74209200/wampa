//go:build medium
// +build medium

package acceptance

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format: "pretty",
			Paths: []string{
				"../../features/local_file_monitoring.feature",
				"../../features/config_file_handling.feature",
			},
			TestingT: t,
			// タグ指定を削除して全てのシナリオが実行されるようにする
			// または明示的にconfig_file_handling.featureのシナリオを含めるタグ式を使用する
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
