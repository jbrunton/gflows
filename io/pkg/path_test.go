package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRemotePath(t *testing.T) {
	assert.True(t, IsRemotePath("http://example.com"))
	assert.True(t, IsRemotePath("https://example.com"))
	assert.False(t, IsRemotePath("/some/local/absolute/path"))
	assert.False(t, IsRemotePath("some/local/relative/path"))
	assert.False(t, IsRemotePath("http/local/dir"))
}

func TestParentPath(t *testing.T) {
	assertParentPath(t, "/path/to", "/path/to/my-file")
	assertParentPath(t, "../relative/path/to", "../relative/path/to/my-file")
	assertParentPath(t, "http://example.com/path/to", "http://example.com/path/to/my-file")
	assertParentPath(t, "https://example.com/path/to", "https://example.com/path/to/my-file")
}

func TestJoinRelativePath(t *testing.T) {
	assertJoinRelativePath(t, "/path/to/my-file", "/path/to", "my-file")
	assertJoinRelativePath(t, "/path/to/my-file", "/path", "to/my-file")
	assertJoinRelativePath(t, "https://example.com/path/to/my-file", "https://example.com/path/to", "my-file")
	assertJoinRelativePath(t, "https://example.com/path/to/my-file", "https://example.com/path", "to/my-file")
}

func assertParentPath(t *testing.T, expected string, path string) {
	actual, err := ParentPath(path)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func assertJoinRelativePath(t *testing.T, expected string, rootPath string, relPath string) {
	actual, err := JoinRelativePath(rootPath, relPath)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
