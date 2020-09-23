package runner

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/cmd"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/fixtures"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/styles"
	"github.com/jbrunton/gflows/workflow/action"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type TestFile struct {
	Path    string
	Content string
	Source  string
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

type TestRunner struct {
	testPath     string
	test         *Test
	useMemFs     bool
	fs           *afero.Afero
	out          *bytes.Buffer
	container    *content.Container
	assert       Assertions
	roundTripper *fixtures.MockRoundTripper
}

func NewTestRunner(osFs *afero.Afero, testPath string, useMemFs bool, assert Assertions) *TestRunner {
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
	roundTripper := fixtures.NewMockRoundTripper()
	ioContainer := io.NewContainer(fs, io.NewLogger(out, false, false), styles.NewStyles(false))
	contentContainer := content.NewContainer(ioContainer, &http.Client{Transport: roundTripper})

	return &TestRunner{
		testPath:     testPath,
		test:         &test,
		useMemFs:     useMemFs,
		out:          out,
		container:    contentContainer,
		assert:       assert,
		fs:           fs,
		roundTripper: roundTripper,
	}
}

func (runner *TestRunner) setup(e2eDirectory string) error {
	projectDirectory := filepath.Dir(e2eDirectory)
	for _, file := range runner.test.Setup.Files {
		content := file.Content
		if file.Source != "" {
			source, err := runner.fs.ReadFile(filepath.Join(projectDirectory, file.Source))
			if err != nil {
				return err
			}
			content = string(source)
		}
		if strings.HasPrefix(file.Path, "http://") || strings.HasPrefix(file.Path, "https://") {
			runner.roundTripper.StubBody(file.Path, content)
		} else {
			err := runner.container.ContentWriter().SafelyWriteFile(file.Path, content)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (runner *TestRunner) Run() {
	fs := runner.container.FileSystem()
	cd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if !runner.useMemFs {
		tmpDir, err := fs.TempDir("", "gflows-e2e")
		if err != nil {
			panic(err)
		}
		defer fs.RemoveAll(tmpDir)

		os.Chdir(tmpDir)
		defer os.Chdir(cd)
	}

	err = runner.setup(cd)
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

func (runner *TestRunner) buildContainer(cmd *cobra.Command) (*action.Container, error) {
	opts := config.CreateContextOpts(cmd)
	opts.EnableColors = false
	context, err := config.NewContext(runner.container.FileSystem(), runner.container.Logger(), opts)
	if err != nil {
		return nil, err
	}

	container := action.NewContainer(runner.container, context)
	return container, nil
}
