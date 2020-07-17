package content

import (
	"strings"
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	_ "github.com/jbrunton/gflows/statik"
	"github.com/stretchr/testify/assert"
)

func TestLogErrors(t *testing.T) {
	container, out := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	writer := NewWriter(container)

	writer.LogErrors("path/to/file", "error message", []string{"error details"})

	expectedOutput := "      error path/to/file error message\n  ► error details\n"
	assert.Equal(t, expectedOutput, out.String())
}

func TestSafelyWriteFile(t *testing.T) {
	container, _ := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	writer := NewWriter(container)

	writer.SafelyWriteFile("path/to/file", "foobar")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
}

func TestUpdateFileContentCreate(t *testing.T) {
	container, out := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	writer := NewWriter(container)

	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "     create path/to/file (baz)\n", out.String())
}

func TestUpdateFileContentUpdate(t *testing.T) {
	container, out := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	writer := NewWriter(container)

	writer.SafelyWriteFile("path/to/file", "foo")
	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "     update path/to/file (baz)\n", out.String())
}

func TestUpdateFileContentIdentical(t *testing.T) {
	container, out := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	writer := NewWriter(container)

	writer.SafelyWriteFile("path/to/file", "foobar")
	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := container.FileSystem().ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "  identical path/to/file (baz)\n", out.String())
}

func TestApplyGenerator(t *testing.T) {
	// arrange
	sourceFs := fixtures.CreateTestFileSystem([]fixtures.File{
		{Name: "foo.txt", Body: "foo"},
		{Name: "bar.txt", Body: "bar"},
	}, "TestApplyGenerator")
	generator := WorkflowGenerator{
		Name:    "foo",
		Sources: []string{"/foo.txt", "/bar.txt"},
	}

	container, out := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	fs := container.FileSystem()
	writer := NewWriter(container)
	writer.SafelyWriteFile(".gflows/bar.txt", "baz")

	// act
	writer.ApplyGenerator(sourceFs, container.Context(), generator)

	// assert
	assert.Equal(t, strings.Join([]string{
		"     create .gflows/foo.txt",
		"     update .gflows/bar.txt\n",
	}, "\n"), out.String())

	fooContent, _ := fs.ReadFile(".gflows/foo.txt")
	assert.Equal(t, "foo", string(fooContent))

	barContent, _ := fs.ReadFile(".gflows/bar.txt")
	assert.Equal(t, "bar", string(barContent))
}
