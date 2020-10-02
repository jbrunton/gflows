package e2e

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"

	"github.com/jbrunton/gflows/e2e/runner"
	"github.com/jbrunton/gflows/io"
	_ "github.com/jbrunton/gflows/static/statik"
)

func runTests(t *testing.T, glob string, useMemFs bool) {
	osFs := io.CreateOsFs()

	testFiles, err := afero.Glob(osFs, glob)
	if err != nil {
		panic(err)
	}

	if len(testFiles) == 0 {
		panic(fmt.Errorf("No test files found: %s", glob))
	}

	for _, testFile := range testFiles {
		assertions := runner.NewTestifyAssertions(t)
		runner := runner.NewTestRunner(osFs, testFile, useMemFs, assertions)
		runner.Run()
	}

	t.Logf("Completed %d tests for %s", len(testFiles), glob)
}

func TestCheckCommand(t *testing.T) {
	runTests(t, "./tests/check/jsonnet/*.yml", true)
	runTests(t, "./tests/check/ytt/*.yml", true)
}

func TestImportCommand(t *testing.T) {
	runTests(t, "./tests/import/jsonnet/*.yml", true)
	runTests(t, "./tests/import/ytt/*.yml", true)
}

func TestInitCommand(t *testing.T) {
	runTests(t, "./tests/init/jsonnet/*.yml", true)
	runTests(t, "./tests/init/ytt/*.yml", true)
	runTests(t, "./tests/init/errors/*.yml", true)
}

func TestListCommand(t *testing.T) {
	runTests(t, "./tests/ls/jsonnet/*.yml", true)
	runTests(t, "./tests/ls/ytt/*.yml", true)
}

func TestUpdateCommand(t *testing.T) {
	runTests(t, "./tests/update/jsonnet/*.yml", true)
	runTests(t, "./tests/update/ytt/*.yml", true)
}

func TestLocalLibs(t *testing.T) {
	runTests(t, "./tests/local-libs/jsonnet/*.yml", false)
	runTests(t, "./tests/local-libs/ytt/*.yml", false)
}

func TestMiscErrors(t *testing.T) {
	runTests(t, "./tests/misc-errors/*.yml", false)
}

func TestGFlowsPkgs(t *testing.T) {
	runTests(t, "./tests/gflowspkgs/jsonnet/*.yml", false)
	runTests(t, "./tests/gflowspkgs/ytt/*", false)
}
