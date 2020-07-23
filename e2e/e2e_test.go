package e2e

import (
	"testing"

	"github.com/spf13/afero"

	"github.com/jbrunton/gflows/adapters"
	_ "github.com/jbrunton/gflows/statik"
)

func runTests(t *testing.T, glob string, useMemFs bool) {
	osFs := adapters.CreateOsFs()

	testFiles, err := afero.Glob(osFs, glob)
	if err != nil {
		panic(err)
	}

	for _, testFile := range testFiles {
		runner := newE2eTestRunner(osFs, testFile, useMemFs)
		runner.run(t)
	}
}

func TestCheckCommand(t *testing.T) {
	runTests(t, "./check/jsonnet/*.yml", true)
	runTests(t, "./check/ytt/*.yml", true)
}

func TestImportCommand(t *testing.T) {
	runTests(t, "./import/jsonnet/*.yml", true)
	runTests(t, "./import/ytt/*.yml", true)
}

func TestInitCommand(t *testing.T) {
	runTests(t, "./init/jsonnet/*.yml", true)
	runTests(t, "./init/ytt/*.yml", true)
	runTests(t, "./init/errors/*.yml", true)
}

func TestListCommand(t *testing.T) {
	runTests(t, "./ls/jsonnet/*.yml", true)
	runTests(t, "./ls/ytt/*.yml", true)
}

func TestUpdateCommand(t *testing.T) {
	runTests(t, "./update/jsonnet/*.yml", true)
	runTests(t, "./update/ytt/*.yml", true)
}

func TestJPath(t *testing.T) {
	runTests(t, "./jpath/*.yml", false)
}

func TestMiscErrors(t *testing.T) {
	runTests(t, "./misc-errors/*.yml", false)
}
