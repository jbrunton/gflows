package env

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/jbrunton/gflows/io/content"
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
	assert.Regexp(t, "my-lib.gflowslib[0-9]+$", lib.LocalDir) // test it's a temp dir
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
	assert.Regexp(t, "my-lib.gflowslib[0-9]+$", lib.LocalDir) // test it's a temp dir
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "lib/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}
