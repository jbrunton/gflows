package env

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/pkg"
)

func newTestLib(manifestPath string) (*GFlowsLib, *content.Container, *fixtures.MockRoundTripper) {
	ioContainer, context, _ := fixtures.NewTestContext("")
	roundTripper := fixtures.NewMockRoundTripper()
	httpClient := &http.Client{Transport: roundTripper}
	container := content.NewContainer(ioContainer, httpClient)
	installer := NewGFlowsLibInstaller(container.FileSystem(), container.ContentReader(), container.ContentWriter(), container.Logger())
	lib := NewGFlowsLib(container.FileSystem(), installer, container.Logger(), manifestPath, context)
	return lib, container, roundTripper
}

func TestSetupLocalLib(t *testing.T) {
	lib, container, _ := newTestLib("/path/to/my-lib.gflowslib")
	fs := container.FileSystem()
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/lib/lib.yml", "foo: bar")

	err := lib.Setup()

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib.gflowslib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "lib/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
	assert.False(t, lib.isRemote(), "expected local lib")
}

func TestSetupRemoteLib(t *testing.T) {
	lib, container, roundTripper := newTestLib("https://example.com/path/to/my-lib.gflowslib")
	fs := container.FileSystem()
	roundTripper.StubBody("https://example.com/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/lib/lib.yml", "foo: bar")

	err := lib.Setup()

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib.gflowslib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "lib/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}

func TestCleanUp(t *testing.T) {
	// arrange
	lib, container, _ := newTestLib("/path/to/my-lib.gflowslib")
	fs := container.FileSystem()
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/lib/lib.yml", "foo: bar")

	err := lib.Setup()
	assert.NoError(t, err)

	exists, err := fs.Exists(lib.LocalDir)
	assert.NoError(t, err)
	assert.True(t, exists, "expected LocalDir to exist")

	// act
	lib.CleanUp()

	// assert
	exists, err = fs.Exists(lib.LocalDir)
	assert.NoError(t, err)
	assert.False(t, exists, "expected LocalDir to have been removed")
}

func TestGetLocalPathInfo(t *testing.T) {
	lib, container, _ := newTestLib("/path/to/my-lib.gflowslib")
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": []}`)
	err := lib.Setup()
	assert.NoError(t, err)

	localPath := filepath.Join(lib.LocalDir, "foo/bar.yml")
	info, err := lib.GetPathInfo(localPath)

	assert.NoError(t, err)
	assert.Equal(t, &pkg.PathInfo{
		LocalPath:   localPath,
		SourcePath:  "/path/to/foo/bar.yml",
		Description: "my-lib.gflowslib/foo/bar.yml",
	}, info)
}

func TestGetRemotePathInfo(t *testing.T) {
	lib, _, roundTripper := newTestLib("https://example.com/path/to/my-lib.gflowslib")
	roundTripper.StubBody("https://example.com/path/to/my-lib.gflowslib", `{"libs": []}`)
	err := lib.Setup()
	assert.NoError(t, err)

	localPath := filepath.Join(lib.LocalDir, "foo/bar.yml")
	info, err := lib.GetPathInfo(localPath)

	assert.NoError(t, err)
	assert.Equal(t, &pkg.PathInfo{
		LocalPath:   localPath,
		SourcePath:  "https://example.com/path/to/foo/bar.yml",
		Description: "my-lib.gflowslib/foo/bar.yml",
	}, info)
}

func TestGetPathInfoErrors(t *testing.T) {
	lib, container, _ := newTestLib("/path/to/my-lib.gflowslib")
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": []}`)
	err := lib.Setup()
	assert.NoError(t, err)

	_, err = lib.GetPathInfo("foo/bar.yml")
	assert.EqualError(t, err, "Expected foo/bar.yml to be absolute")

	_, err = lib.GetPathInfo("/path")
	assert.Regexp(t, fmt.Sprintf("^Expected /path to be a subdirectory of %s", lib.LocalDir), err)
}
