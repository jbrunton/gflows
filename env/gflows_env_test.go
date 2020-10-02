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
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"files": ["libs/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/libs/lib.yml", "foo: bar")

	lib, err := env.LoadDependency("/path/to/my-lib.gflowslib")

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib.gflowslib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
	assert.False(t, lib.isRemote(), "expected local lib")
}

func TestLoadRemoteLib(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	env, container := newTestEnv(roundTripper)
	fs := container.FileSystem()
	roundTripper.StubBody("https://example.com/path/to/my-lib.gflowslib", `{"files": ["libs/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/libs/lib.yml", "foo: bar")

	lib, err := env.LoadDependency("https://example.com/path/to/my-lib.gflowslib")

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib.gflowslib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
	assert.True(t, lib.isRemote(), "expected remote lib")
}

func TestCacheRemoteLibs(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	env, _ := newTestEnv(roundTripper)
	roundTripper.StubBody("https://example.com/path/to/my-lib.gflowslib", `{"files": ["libs/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/libs/lib.yml", "foo: bar")

	libOne, err := env.LoadDependency("https://example.com/path/to/my-lib.gflowslib")
	assert.NoError(t, err)
	libTwo, err := env.LoadDependency("https://example.com/path/to/my-lib.gflowslib")
	assert.NoError(t, err)

	assert.True(t, libOne == libTwo, "expected same lib")
	roundTripper.AssertNumberOfCalls(t, "RoundTrip", 2) // one call for the manifest and another for the lib.yml file
	roundTripper.AssertExpectations(t)
}

func TestGetPackages(t *testing.T) {
	// arrange
	env, container := newTestEnv(fixtures.NewMockRoundTripper())
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"files": ["libs/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/libs/lib.yml", "foo: bar")
	lib, _ := env.LoadDependency("/path/to/my-lib.gflowslib")

	// act
	packages, err := env.GetPackages()

	// assert
	libPackage := packages[0]
	contextPackage := packages[1]
	assert.Len(t, packages, 2)
	assert.NoError(t, err)

	assert.Equal(t, ".gflows/workflows", contextPackage.WorkflowsDir())
	assert.Equal(t, ".gflows/libs", contextPackage.LibsDir())

	assert.Equal(t, filepath.Join(lib.LocalDir, "workflows"), libPackage.WorkflowsDir())
	assert.Equal(t, filepath.Join(lib.LocalDir, "libs"), libPackage.LibsDir())
}

// func TestGetJPath(t *testing.T) {
// 	config := strings.Join([]string{
// 		"templates:",
// 		"  engine: jsonnet",
// 		"  defaults:",
// 		"    libs: [some-lib]",
// 		"  overrides:",
// 		"    my-workflow:",
// 		"      libs: [my-lib]",
// 	}, "\n")
// 	_, _, engine := newJsonnetTemplateEngine(config, fixtures.NewMockRoundTripper())

// 	jpath, _ := engine.getJPath("my-workflow")
// 	assert.Equal(t, []string{".gflows/some-lib", ".gflows/my-lib"}, jpath)

// 	jpath, _ = engine.getJPath("other-workflow")
// 	assert.Equal(t, []string{".gflows/some-lib"}, jpath)
// }

// func TestJPathErrors(t *testing.T) {
// 	roundTripper := fixtures.NewMockRoundTripper()
// 	roundTripper.StubStatusCode("https://example.com/my-lib.gflowslib", 500)
// 	config := strings.Join([]string{
// 		"templates:",
// 		"  engine: jsonnet",
// 		"  defaults:",
// 		"    libs: [https://example.com/my-lib.gflowslib]",
// 	}, "\n")
// 	_, _, engine := newJsonnetTemplateEngine(config, roundTripper)

// 	_, err := engine.getJPath("my-workflow")
// 	assert.EqualError(t, err, "Received status code 500 from https://example.com/my-lib.gflowslib")
// }

// func TestGetYttLibs(t *testing.T) {
// 	config := strings.Join([]string{
// 		"templates:",
// 		"  engine: ytt",
// 		"  defaults:",
// 		"    libs: [common, config]",
// 		"  overrides:",
// 		"    my-workflow:",
// 		"      libs: [my-lib]",
// 	}, "\n")
// 	_, _, engine, _ := newYttTemplateEngine(config)

// 	paths, err := engine.getYttLibs("my-workflow")
// 	assert.NoError(t, err)
// 	assert.Equal(t, []string{".gflows/common", ".gflows/config", ".gflows/my-lib"}, paths)

// 	paths, err = engine.getYttLibs("other-workflow")
// 	assert.NoError(t, err)
// 	assert.Equal(t, []string{".gflows/common", ".gflows/config"}, paths)
// }

// func TestRemoteLibs(t *testing.T) {
// 	config := strings.Join([]string{
// 		"templates:",
// 		"  engine: ytt",
// 		"  defaults:",
// 		"    libs: [https://example.com/my-lib.gflowslib]",
// 		"  overrides:",
// 		"    my-workflow:",
// 		"      libs: [https://example.com/other-lib.gflowslib]",
// 	}, "\n")
// 	_, _, engine, roundTripper := newYttTemplateEngine(config)
// 	roundTripper.StubBody("https://example.com/my-lib.gflowslib", `{"files":[]}`)
// 	roundTripper.StubBody("https://example.com/other-lib.gflowslib", `{"files":[]}`)

// 	paths, err := engine.getYttLibs("my-workflow")
// 	assert.NoError(t, err)
// 	assert.Equal(t, len(paths), 2)
// 	assert.Regexp(t, "my-lib.gflowslib[0-9]+/libs$", paths[0])
// 	assert.Regexp(t, "other-lib.gflowslib[0-9]+/libs$", paths[1])
// }
