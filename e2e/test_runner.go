package e2e

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/cmd"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/styles"
	"github.com/jbrunton/gflows/workflows"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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
	testPath  string
	test      *e2eTest
	useMemFs  bool
	out       *bytes.Buffer
	container *content.Container
}

func newE2eTestRunner(osFs *afero.Afero, testPath string, useMemFs bool) *e2eTestRunner {
	test := e2eTest{}
	input, err := osFs.ReadFile(testPath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(input, &test)
	if err != nil {
		panic(err)
	}

	var fs *afero.Afero
	if useMemFs {
		fs = adapters.CreateMemFs()
	} else {
		fs = osFs
	}

	out := new(bytes.Buffer)
	adaptersContainer := adapters.NewContainer(fs, adapters.NewLogger(out), styles.NewStyles(false))
	contentContainer := content.NewContainer(adaptersContainer)

	return &e2eTestRunner{
		testPath:  testPath,
		test:      &test,
		useMemFs:  useMemFs,
		out:       out,
		container: contentContainer,
	}
}

func (runner *e2eTestRunner) setup() error {
	for _, file := range runner.test.Setup.Files {
		err := runner.container.ContentWriter().SafelyWriteFile(file.Path, file.Content)
		if err != nil {
			return err
		}
	}
	return nil
}

func (runner *e2eTestRunner) run(t *testing.T) {
	fs := runner.container.FileSystem()
	if !runner.useMemFs {
		tmpDir, err := fs.TempDir("", "gflows-e2e")
		if err != nil {
			panic(err)
		}
		defer fs.RemoveAll(tmpDir)

		cd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		os.Chdir(tmpDir)
		defer os.Chdir(cd)
	}

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
			exists, err := fs.Exists(expectedFile.Path)
			if err != nil {
				panic(err)
			}
			assert.True(t, exists, "Expected file %s to exist (%s)", expectedFile.Path, runner.testPath)
			if exists && expectedFile.Content != "" {
				actualContent, err := fs.ReadFile(expectedFile.Path)
				if err != nil {
					panic(err)
				}
				assert.Equal(t, expectedFile.Content, string(actualContent), "Unexpected content for file %s (%s)", expectedFile.Path, runner.testPath)
			}
		}
		fs.Walk(".", func(path string, info os.FileInfo, err error) error {
			dir, err := fs.IsDir(path)
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
	context, err := config.GetContext(runner.container.FileSystem(), cmd)
	context.EnableColors = false
	if err != nil {
		return nil, err
	}

	container := workflows.NewContainer(runner.container, context)
	return container, nil
}
