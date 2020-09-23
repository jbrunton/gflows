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

func newTestEnv() (*GFlowsEnv, *io.Container, *fixtures.TestRoundTripper) {
	container, context, out := fixtures.NewTestContext("")
	roundTripper := fixtures.NewTestRoundTripper()
	fs := container.FileSystem()
	logger := io.NewLogger(out, false, false)
	writer := content.NewWriter(fs, logger)
	httpClient := &http.Client{Transport: roundTripper}
	downloader := content.NewDownloader(fs, writer, httpClient, logger)
	env := NewGFlowsEnv(fs, downloader, context, logger)
	return env, container, roundTripper
}

func TestLoadLocalLibrary(t *testing.T) {
	env, _, _ := newTestEnv()

	lib, err := env.LoadLib("/path/to/my-lib.gflowslib")

	assert.NoError(t, err)
	assert.Equal(t, "/path/to", lib.LocalDir)
	assert.False(t, lib.isRemote(), "expected local lib")
}

func TestLoadRemoteLib(t *testing.T) {
	env, container, roundTripper := newTestEnv()
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
