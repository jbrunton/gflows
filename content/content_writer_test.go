package content

import (
	"strings"
	"testing"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/fixtures"
	_ "github.com/jbrunton/gflows/statik"
	"github.com/stretchr/testify/assert"
)

func TestLogErrors(t *testing.T) {
	fs, _, out := fixtures.NewTestContext("")
	writer := NewWriter(fs, adapters.NewLogger(out))

	writer.LogErrors("path/to/file", "error message", []string{"error details"})

	expectedOutput := "      error path/to/file error message\n  ► error details\n"
	assert.Equal(t, expectedOutput, out.String())
}

func TestSafelyWriteFile(t *testing.T) {
	fs, _, out := fixtures.NewTestContext("")
	writer := NewWriter(fs, adapters.NewLogger(out))

	writer.SafelyWriteFile("path/to/file", "foobar")

	actualContent, _ := fs.ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
}

func TestUpdateFileContentCreate(t *testing.T) {
	fs, _, out := fixtures.NewTestContext("")
	writer := NewWriter(fs, adapters.NewLogger(out))

	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := fs.ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "     create path/to/file (baz)\n", out.String())
}

func TestUpdateFileContentUpdate(t *testing.T) {
	fs, _, out := fixtures.NewTestContext("")
	writer := NewWriter(fs, adapters.NewLogger(out))

	writer.SafelyWriteFile("path/to/file", "foo")
	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := fs.ReadFile("path/to/file")
	assert.Equal(t, "foobar", string(actualContent))
	assert.Equal(t, "     update path/to/file (baz)\n", out.String())
}

func TestUpdateFileContentIdentical(t *testing.T) {
	fs, _, out := fixtures.NewTestContext("")
	writer := NewWriter(fs, adapters.NewLogger(out))

	writer.SafelyWriteFile("path/to/file", "foobar")
	writer.UpdateFileContent("path/to/file", "foobar", "(baz)")

	actualContent, _ := fs.ReadFile("path/to/file")
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

	fs, context, out := fixtures.NewTestContext("")
	writer := NewWriter(fs, adapters.NewLogger(out))
	writer.SafelyWriteFile(".gflows/bar.txt", "baz")

	// act
	writer.ApplyGenerator(sourceFs, context, generator)

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
