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
	runTests(t, "./check/*.yml", true)
}

func TestImportCommand(t *testing.T) {
	runTests(t, "./import/*.yml", true)
}

func TestInitCommand(t *testing.T) {
	runTests(t, "./init/*.yml", true)
}

func TestListCommand(t *testing.T) {
	runTests(t, "./ls/*.yml", true)
}

func TestUpdateCommand(t *testing.T) {
	runTests(t, "./update/*.yml", true)
}

func TestJPath(t *testing.T) {
	runTests(t, "./jpath/*.yml", false)
}
