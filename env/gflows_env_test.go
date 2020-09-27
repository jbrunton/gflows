package env

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/jbrunton/gflows/io/content"
)

func newTestEnv(roundTripper http.RoundTripper) (*GFlowsEnv, *content.Container) {
	ioContainer, context, _ := fixtures.NewTestContext("")
	httpClient := &http.Client{Transport: roundTripper}
	container := content.NewContainer(ioContainer, httpClient)
	installer := NewGFlowsLibInstaller(container.FileSystem(), container.ContentReader(), container.ContentWriter(), container.Logger())
	env := NewGFlowsEnv(container.FileSystem(), installer, context, container.Logger())
	return env, container
}

func TestLoadLocalLibrary(t *testing.T) {
	env, container := newTestEnv(fixtures.NewMockRoundTripper())
	fs := container.FileSystem()
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/lib/lib.yml", "foo: bar")

	lib, err := env.LoadLib("/path/to/my-lib.gflowslib")

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib.gflowslib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "lib/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
	assert.False(t, lib.isRemote(), "expected local lib")
}

func TestLoadRemoteLib(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	env, container := newTestEnv(roundTripper)
	fs := container.FileSystem()
	roundTripper.StubBody("https://example.com/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/lib/lib.yml", "foo: bar")

	lib, err := env.LoadLib("https://example.com/path/to/my-lib.gflowslib")

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib.gflowslib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "lib/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
	assert.True(t, lib.isRemote(), "expected remote lib")
}

func TestCacheRemoteLibs(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	env, _ := newTestEnv(roundTripper)
	roundTripper.StubBody("https://example.com/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/lib/lib.yml", "foo: bar")

	libOne, err := env.LoadLib("https://example.com/path/to/my-lib.gflowslib")
	assert.NoError(t, err)
	libTwo, err := env.LoadLib("https://example.com/path/to/my-lib.gflowslib")
	assert.NoError(t, err)

	assert.True(t, libOne == libTwo, "expected same lib")
	roundTripper.AssertNumberOfCalls(t, "RoundTrip", 2) // one call for the manifest and another for the lib.yml file
	roundTripper.AssertExpectations(t)
}

func TestGetLibPaths(t *testing.T) {
	env, container := newTestEnv(fixtures.NewMockRoundTripper())
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/lib/lib.yml", "foo: bar")
	lib, _ := env.LoadLib("/path/to/my-lib.gflowslib")

	libDirs := env.GetLibDirs()

	assert.Len(t, libDirs, 2)
	assert.Equal(t, ".gflows/libs", libDirs[0])
	assert.Equal(t, filepath.Join(lib.LocalDir, "libs"), libDirs[1])
}

func TestGetWorkflowPaths(t *testing.T) {
	env, container := newTestEnv(fixtures.NewMockRoundTripper())
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/lib/lib.yml", "foo: bar")
	lib, _ := env.LoadLib("/path/to/my-lib.gflowslib")

	libDirs := env.GetWorkflowDirs()

	assert.Len(t, libDirs, 2)
	assert.Equal(t, ".gflows/workflows", libDirs[0])
	assert.Equal(t, filepath.Join(lib.LocalDir, "workflows"), libDirs[1])
}
