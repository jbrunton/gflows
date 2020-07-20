package e2e

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jbrunton/gflows/styles"

	"github.com/jbrunton/gflows/cmd"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/workflows"

	"github.com/stretchr/testify/assert"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/jbrunton/gflows/adapters"
	_ "github.com/jbrunton/gflows/statik"
)

type e2eTestFile struct {
	Path    string
	Content string
}

type e2eTestSetup struct {
	Files []e2eTestFile
}

type e2eTestExpect struct {
	Output string
	Error  string
	Files  []e2eTestFile
}

type e2eTest struct {
	Run    string
	Setup  e2eTestSetup
	Expect e2eTestExpect
}

type e2eTestRunner struct {
	testPath string
	out      *bytes.Buffer
	fs       *afero.Afero
	test     *e2eTest
}

func newE2eTestRunner(testPath string, test *e2eTest) *e2eTestRunner {
	return &e2eTestRunner{
		testPath: testPath,
		out:      new(bytes.Buffer),
		fs:       adapters.CreateMemFs(),
		test:     test,
	}
}

func (runner *e2eTestRunner) setup() error {
	for _, file := range runner.test.Setup.Files {
		err := runner.fs.WriteFile(file.Path, []byte(file.Content), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func (runner *e2eTestRunner) run(t *testing.T) {
	err := runner.setup()
	if err != nil {
		panic(err)
	}

	cmd := cmd.NewRootCommand(runner.buildContainer)
	cmd.SetArgs(strings.Split(runner.test.Run, " "))
	err = cmd.Execute()

	if runner.test.Expect.Error == "" {
		assert.NoError(t, err, "Unexpected error (%s)", runner.testPath)
	} else {
		assert.EqualError(t, err, runner.test.Expect.Error, "Unexpected error (%s)", runner.testPath)
	}
	assert.Equal(t, runner.test.Expect.Output, runner.out.String(), "Unexpected output (%s)", runner.testPath)
	if len(runner.test.Expect.Files) > 0 {
		for _, expectedFile := range runner.test.Expect.Files {
			exists, err := runner.fs.Exists(expectedFile.Path)
			if err != nil {
				panic(err)
			}
			assert.True(t, exists, "Expected file %s to exist (%s)", expectedFile.Path, runner.testPath)
			if exists && expectedFile.Content != "" {
				actualContent, err := runner.fs.ReadFile(expectedFile.Path)
				if err != nil {
					panic(err)
				}
				assert.Equal(t, expectedFile.Content, string(actualContent), "Unexpected content for file %s (%s)", expectedFile.Path, runner.testPath)
			}
		}
		runner.fs.Walk(".", func(path string, info os.FileInfo, err error) error {
			dir, err := runner.fs.IsDir(path)
			if err != nil {
				panic(err)
			}
			if dir {
				return nil
			}
			expected := false
			for _, expectedFile := range runner.test.Expect.Files {
				if expectedFile.Path == path {
					expected = true
				}
			}
			assert.True(t, expected, "File %s was not expected (%s)", path, runner.testPath)
			return nil
		})
	}
}

func (runner *e2eTestRunner) buildContainer(cmd *cobra.Command) (*workflows.Container, error) {
	adaptersContainer := adapters.NewContainer(runner.fs, adapters.NewLogger(runner.out), styles.NewStyles(false))
	contentContainer := content.NewContainer(adaptersContainer)

	context, err := config.GetContext(adaptersContainer.FileSystem(), cmd)
	context.EnableColors = false
	if err != nil {
		return nil, err
	}

	return workflows.NewContainer(contentContainer, context), nil
}

func TestE2e(t *testing.T) {
	osFs := adapters.CreateOsFs()

	testFiles, err := afero.Glob(osFs, "./*/*.yml")
	if err != nil {
		panic(err)
	}

	for _, testFile := range testFiles {
		fmt.Printf("Starting test run for %s\n", testFile)

		test := e2eTest{}
		input, err := osFs.ReadFile(testFile)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(input, &test)
		if err != nil {
			panic(err)
		}

		runner := newE2eTestRunner(testFile, &test)
		runner.run(t)
	}
}
