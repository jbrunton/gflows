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

func newTestLib(manifestPath string) (*GFlowsLib, *io.Container, *fixtures.TestRoundTripper) {
	container, context, out := fixtures.NewTestContext("")
	roundTripper := fixtures.NewTestRoundTripper()
	fs := container.FileSystem()
	logger := io.NewLogger(out, false, false)
	writer := content.NewWriter(fs, logger)
	httpClient := &http.Client{Transport: roundTripper}
	downloader := content.NewDownloader(fs, writer, httpClient, logger)
	lib := NewGFlowsLib(fs, downloader, logger, manifestPath, context)
	return lib, container, roundTripper
}

func TestSetupLocalLib(t *testing.T) {
	lib, _, _ := newTestLib("/path/to/my-lib.gflowslib")

	err := lib.Setup()

	assert.NoError(t, err)
	assert.Equal(t, "/path/to", lib.LocalDir)
}

func TestSetupRemoteLib(t *testing.T) {
	lib, container, roundTripper := newTestLib("https://example.com/path/to/my-lib.gflowslib")
	fs := container.FileSystem()
	roundTripper.StubBody("https://example.com/path/to/my-lib.gflowslib", `{"libs": ["lib/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/lib/lib.yml", "foo: bar")

	err := lib.Setup()

	assert.NoError(t, err)
	assert.Regexp(t, "^/var/folders.*my-lib.gflowslib[0-9]+$", lib.LocalDir) // test it's a temp dir
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "lib/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}
