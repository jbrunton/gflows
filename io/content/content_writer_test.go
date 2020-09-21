package content

import (
	"strings"
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	_ "github.com/jbrunton/gflows/static/statik"
	"github.com/stretchr/testify/assert"
)

func TestLogErrors(t *testing.T) {
	container, _, out := fixtures.NewTestContext("")
	writer := NewWriter(container.FileSystem(), container.Logger())

	writer.LogErrors("path/to/file", "error message", []string{"error details"})

	expectedOutput := "      error path/to/file error message\n  ► error details\n"
	assert.Equal(t, expectedOutput, out.String())
}

func TestSafelyWriteFile(t *testing.T) {
	container, _, _ := fixtures.NewTestContext("")
	writer := NewWriter(container.FileSystem(), container.Logger())

	writer.SafelyWriteFile("path/to/file", "foobar")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
}

func TestUpdateFileContentCreate(t *testing.T) {
	container, _, out := fixtures.NewTestContext("")
	writer := NewWriter(container.FileSystem(), container.Logger())

	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "     create path/to/file (baz)\n", out.String())
}

func TestUpdateFileContentUpdate(t *testing.T) {
	container, _, out := fixtures.NewTestContext("")
	writer := NewWriter(container.FileSystem(), container.Logger())

	writer.SafelyWriteFile("path/to/file", "foo")
	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "     update path/to/file (baz)\n", out.String())
}

func TestUpdateFileContentIdentical(t *testing.T) {
	container, _, out := fixtures.NewTestContext("")
	writer := NewWriter(container.FileSystem(), container.Logger())

	writer.SafelyWriteFile("path/to/file", "foobar")
	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "  identical path/to/file (baz)\n", out.String())
}

func TestApplyGenerator(t *testing.T) {
	// arrange
	sourceFs := fixtures.CreateTestFileSystem([]fixtures.File{
		{Path: "foo.txt", Content: "foo"},
		{Path: "jsonnet/bar.txt", Content: "bar"},
	}, "TestApplyGenerator")
	generator := WorkflowGenerator{
		Name: "foo",
		TemplateVars: map[string]string{
			"WORKFLOW_NAME": "gflows",
			"JOB_NAME":      "check-workflows",
			"GITHUB_DIR":    ".github",
			"CONFIG_PATH":   ".gflows/config.yml",
		},
		Sources: []WorkflowSource{
			NewWorkflowSource("/foo.txt", "/foo.txt"),
			NewWorkflowSource("/jsonnet/bar.txt", "/bar.txt"),
		},
	}

	container, context, out := fixtures.NewTestContext("")
	writer := NewWriter(container.FileSystem(), container.Logger())
	writer.SafelyWriteFile(".gflows/bar.txt", "baz")

	// act
	writer.ApplyGenerator(sourceFs, context.Dir, generator)

	// assert
	assert.Equal(t, strings.Join([]string{
		"     create .gflows/foo.txt",
		"     update .gflows/bar.txt\n",
	}, "\n"), out.String())

	fooContent, _ := container.FileSystem().ReadFile(".gflows/foo.txt")
	assert.Equal(t, "foo", string(fooContent))

	barContent, _ := container.FileSystem().ReadFile(".gflows/bar.txt")
	assert.Equal(t, "bar", string(barContent))
}
