package e2e

import (
	"bytes"
	"os"
	"strings"

	"github.com/jbrunton/gflows/cmd"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/styles"
	"github.com/jbrunton/gflows/workflow/action"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

type TestFile struct {
	Path    string
	Content string
}

type TestSetup struct {
	Files []TestFile
}

type TestExpect struct {
	Output string
	Error  string
	Files  []TestFile
}

type Test struct {
	Run    string
	Setup  TestSetup
	Expect TestExpect
}

type Assertions interface {
	NoError(err error, msgAndArgs ...interface{})
	EqualError(theError error, errString string, msgAndArgs ...interface{})
	True(value bool, msgAndArgs ...interface{})
	Equal(expected, actual interface{}, msgAndArgs ...interface{})
}

type TestiftyAssertions struct {
	t assert.TestingT
}

func (a *TestiftyAssertions) NoError(err error, msgAndArgs ...interface{}) {
	assert.NoError(a.t, err, msgAndArgs...)
}

func (a *TestiftyAssertions) EqualError(theError error, errString string, msgAndArgs ...interface{}) {
	assert.EqualError(a.t, theError, errString, msgAndArgs...)
}

func (a *TestiftyAssertions) True(value bool, msgAndArgs ...interface{}) {
	assert.True(a.t, value, msgAndArgs...)
}
func (a *TestiftyAssertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.Equal(a.t, expected, actual, msgAndArgs...)
}

type e2eTestRunner struct {
	testPath  string
	test      *Test
	useMemFs  bool
	out       *bytes.Buffer
	container *content.Container
	assert    Assertions
}

func newE2eTestRunner(osFs *afero.Afero, testPath string, useMemFs bool, assert Assertions) *e2eTestRunner {
	test := Test{}
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
		fs = io.CreateMemFs()
	} else {
		fs = osFs
	}

	out := new(bytes.Buffer)
	ioContainer := io.NewContainer(fs, io.NewLogger(out, false), styles.NewStyles(false))
	contentContainer := content.NewContainer(ioContainer)

	return &e2eTestRunner{
		testPath:  testPath,
		test:      &test,
		useMemFs:  useMemFs,
		out:       out,
		container: contentContainer,
		assert:    assert,
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

func (runner *e2eTestRunner) run() {
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
	args := strings.Split(runner.test.Run, " ")
	cmd.SetArgs(args)
	err = cmd.Execute()

	if runner.test.Expect.Error == "" {
		runner.assert.NoError(err, "Unexpected error (%s)", runner.testPath)
	} else {
		runner.assert.EqualError(err, runner.test.Expect.Error, "Unexpected error (%s)", runner.testPath)
	}
	runner.assert.Equal(runner.test.Expect.Output, runner.out.String(), "Unexpected output (%s)", runner.testPath)
	if len(runner.test.Expect.Files) > 0 {
		for _, expectedFile := range runner.test.Expect.Files {
			exists, err := fs.Exists(expectedFile.Path)
			if err != nil {
				panic(err)
			}
			runner.assert.True(exists, "Expected file %s to exist (%s)", expectedFile.Path, runner.testPath)
			if exists && expectedFile.Content != "" {
				actualContent, err := fs.ReadFile(expectedFile.Path)
				if err != nil {
					panic(err)
				}
				runner.assert.Equal(expectedFile.Content, string(actualContent), "Unexpected content for file %s (%s)", expectedFile.Path, runner.testPath)
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
			runner.assert.True(expected, "File %s was not expected (%s)", path, runner.testPath)
			return nil
		})
	}
}

func (runner *e2eTestRunner) buildContainer(cmd *cobra.Command) (*action.Container, error) {
	opts := config.CreateContextOpts(cmd)
	opts.EnableColors = false
	context, err := config.NewContext(runner.container.FileSystem(), runner.container.Logger(), opts)
	if err != nil {
		return nil, err
	}

	container := action.NewContainer(runner.container, context)
	return container, nil
}
