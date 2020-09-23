package env

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
)

func newTestEnv(roundTripper http.RoundTripper) (*GFlowsEnv, *io.Container) {
	container, context, out := fixtures.NewTestContext("")
	fs := container.FileSystem()
	logger := io.NewLogger(out, false, false)
	writer := content.NewWriter(fs, logger)
	httpClient := &http.Client{Transport: roundTripper}
	downloader := content.NewDownloader(fs, writer, httpClient, logger)
	env := NewGFlowsEnv(fs, downloader, context, logger)
	return env, container
}

func TestLoadLocalLibrary(t *testing.T) {
	env, _ := newTestEnv(fixtures.NewMockRoundTripper())

	lib, err := env.LoadLib("/path/to/my-lib.gflowslib")

	assert.NoError(t, err)
	assert.Equal(t, "/path/to", lib.LocalDir)
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
	assert.Regexp(t, "my-lib.gflowslib[0-9]+$", lib.LocalDir) // test it's a temp dir
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
