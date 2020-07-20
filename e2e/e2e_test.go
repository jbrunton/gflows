package e2e

import (
	"bytes"
	"fmt"
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
		assert.NoError(t, err)
	} else {
		assert.EqualError(t, err, runner.test.Expect.Error)
	}
	assert.Equal(t, strings.TrimSpace(runner.test.Expect.Output), strings.TrimSpace(runner.out.String()))
}

func (runner *e2eTestRunner) buildContainer(cmd *cobra.Command) (*workflows.Container, error) {
	adaptersContainer := adapters.NewContainer(runner.fs, adapters.NewLogger(runner.out), styles.NewStyles(false))
	contentContainer := content.NewContainer(adaptersContainer)

	context, err := config.GetContext(adaptersContainer.FileSystem(), cmd)
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
