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
	repoManager := content.NewRepoManager(container.GitAdapter(), container.FileSystem(), container.Logger())
	installer := NewGFlowsLibInstaller(container.FileSystem(), container.ContentReader(), container.ContentWriter(), container.Logger(), repoManager)
	lib, _ := NewGFlowsLib(container.FileSystem(), installer, container.Logger(), manifestPath, context)
	return lib, container, roundTripper
}

func TestSetupLocalLib(t *testing.T) {
	lib, container, _ := newTestLib("/path/to/my-lib")
	fs := container.FileSystem()
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": ["libs/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/libs/lib.yml", "foo: bar")

	err := lib.Setup()

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}

func TestSetupRemoteLib(t *testing.T) {
	lib, container, roundTripper := newTestLib("https://example.com/path/to/my-lib")
	fs := container.FileSystem()
	roundTripper.StubBody("https://example.com/path/to/my-lib/gflowspkg.json", `{"files": ["libs/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/my-lib/libs/lib.yml", "foo: bar")

	err := lib.Setup()

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}

func TestSetupGitLib(t *testing.T) {
	lib, container, _ := newTestLib("git@example.com:my/repo.git")
	fs := container.FileSystem()
	gitAdapter := container.GitAdapter().(*fixtures.TestGitAdapter)
	gitAdapter.StubRepo("git@example.com:my/repo.git", &map[string]string{
		"gflowspkg.json": `{"files": ["libs/lib.yml"]}`,
		"libs/lib.yml":   "foo: bar",
	})

	err := lib.Setup()

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "repo.git", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}

func TestSetupNestedGitLib(t *testing.T) {
	lib, container, _ := newTestLib("git@example.com:my/repo.git/my-lib")
	fs := container.FileSystem()
	gitAdapter := container.GitAdapter().(*fixtures.TestGitAdapter)
	gitAdapter.StubRepo("git@example.com:my/repo.git", &map[string]string{
		"my-lib/gflowspkg.json": `{"files": ["libs/lib.yml"]}`,
		"my-lib/libs/lib.yml":   "foo: bar",
	})

	err := lib.Setup()

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}

func TestLibStructureErrors(t *testing.T) {
	lib, container, _ := newTestLib("/path/to/my-lib")
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": ["foo/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/foo/lib.yml", "foo: bar")

	err := lib.Setup()

	assert.EqualError(t, err, "Unexpected directory foo/lib.yml, file must be in libs/ or workflows/")
}

func TestCleanUpLib(t *testing.T) {
	// arrange
	lib, container, _ := newTestLib("/path/to/my-lib")
	fs := container.FileSystem()
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": ["libs/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/libs/lib.yml", "foo: bar")

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
	lib, container, _ := newTestLib("/path/to/my-lib")
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": []}`)
	err := lib.Setup()
	assert.NoError(t, err)

	localPath := filepath.Join(lib.LocalDir, "foo/bar.yml")
	info, err := lib.GetPathInfo(localPath)

	assert.NoError(t, err)
	assert.Equal(t, &pkg.PathInfo{
		LocalPath:   localPath,
		SourcePath:  "/path/to/my-lib/foo/bar.yml",
		Description: "my-lib/foo/bar.yml",
	}, info)
}

func TestGetRemotePathInfo(t *testing.T) {
	lib, _, roundTripper := newTestLib("https://example.com/path/to/my-lib")
	roundTripper.StubBody("https://example.com/path/to/my-lib/gflowspkg.json", `{"files": []}`)
	err := lib.Setup()
	assert.NoError(t, err)

	localPath := filepath.Join(lib.LocalDir, "foo/bar.yml")
	info, err := lib.GetPathInfo(localPath)

	assert.NoError(t, err)
	assert.Equal(t, &pkg.PathInfo{
		LocalPath:   localPath,
		SourcePath:  "https://example.com/path/to/my-lib/foo/bar.yml",
		Description: "my-lib/foo/bar.yml",
	}, info)
}

func TestGetPathInfoErrors(t *testing.T) {
	lib, container, _ := newTestLib("/path/to/my-lib")
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": []}`)
	err := lib.Setup()
	assert.NoError(t, err)

	_, err = lib.GetPathInfo("foo/bar.yml")
	assert.EqualError(t, err, "Expected foo/bar.yml to be absolute")

	_, err = lib.GetPathInfo("/path")
	assert.Regexp(t, fmt.Sprintf("^Expected /path to be a subdirectory of %s", lib.LocalDir), err)
}
