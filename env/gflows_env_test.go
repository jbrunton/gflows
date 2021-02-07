package env

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/jbrunton/gflows/io/content"
)

func newTestEnv(config string, roundTripper http.RoundTripper) (*GFlowsEnv, *content.Container) {
	ioContainer, context, _ := fixtures.NewTestContext(config)
	httpClient := &http.Client{Transport: roundTripper}
	container := content.NewContainer(ioContainer, httpClient)
	repoManager := content.NewRepoManager(container.GitAdapter(), container.FileSystem(), container.Logger())
	installer := NewGFlowsLibInstaller(container.FileSystem(), container.ContentReader(), container.ContentWriter(), container.Logger(), repoManager)
	env := NewGFlowsEnv(container.FileSystem(), installer, context, container.Logger())
	return env, container
}

func TestLoadLocalLibrary(t *testing.T) {
	env, container := newTestEnv("", fixtures.NewMockRoundTripper())
	fs := container.FileSystem()
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": ["libs/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/libs/lib.yml", "foo: bar")

	lib, err := env.LoadDependency("/path/to/my-lib")

	assert.NoError(t, err)
	fixtures.AssertTempDir(t, fs, "my-lib", lib.LocalDir)
	assert.Equal(t, "my-lib", lib.PackageName)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}

func TestLoadRemoteLib(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	env, container := newTestEnv("", roundTripper)
	fs := container.FileSystem()
	roundTripper.StubBody("https://example.com/path/to/my-lib/gflowspkg.json", `{"files": ["libs/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/my-lib/libs/lib.yml", "foo: bar")

	lib, err := env.LoadDependency("https://example.com/path/to/my-lib")

	assert.NoError(t, err)
	assert.Equal(t, "my-lib", lib.PackageName)
	fixtures.AssertTempDir(t, fs, "my-lib", lib.LocalDir)
	libContent, _ := fs.ReadFile(filepath.Join(lib.LocalDir, "libs/lib.yml"))
	assert.Equal(t, "foo: bar", string(libContent))
}

func TestCacheRemoteLibs(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	env, _ := newTestEnv("", roundTripper)
	roundTripper.StubBody("https://example.com/path/to/my-lib/gflowspkg.json", `{"files": ["libs/lib.yml"]}`)
	roundTripper.StubBody("https://example.com/path/to/my-lib/libs/lib.yml", "foo: bar")

	libOne, err := env.LoadDependency("https://example.com/path/to/my-lib")
	assert.NoError(t, err)
	libTwo, err := env.LoadDependency("https://example.com/path/to/my-lib")
	assert.NoError(t, err)

	assert.True(t, libOne == libTwo, "expected same lib")
	roundTripper.AssertNumberOfCalls(t, "RoundTrip", 2) // one call for the manifest and another for the lib.yml file
	roundTripper.AssertExpectations(t)
}

func TestCustomPackageName(t *testing.T) {
	env, container := newTestEnv("", fixtures.NewMockRoundTripper())
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": [], "name": "my-lib-name"}`)

	lib, err := env.LoadDependency("/path/to/my-lib")

	assert.NoError(t, err)
	assert.Equal(t, "my-lib-name", lib.PackageName)
}

func TestGetPackages(t *testing.T) {
	// arrange
	env, container := newTestEnv("", fixtures.NewMockRoundTripper())
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/gflowspkg.json", `{"files": ["libs/lib.yml"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib/libs/lib.yml", "foo: bar")
	lib, _ := env.LoadDependency("/path/to/my-lib")

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

func TestGetLibPaths(t *testing.T) {
	config := strings.Join([]string{
		"templates:",
		"  engine: jsonnet",
		"  defaults:",
		"    libs: [/libs/some-lib]",
		"    dependencies: [/deps/some-pkg]",
		"  overrides:",
		"    my-workflow:",
		"      libs: [/libs/my-lib]",
		"      dependencies: [/deps/my-pkg]",
	}, "\n")
	env, container := newTestEnv(config, fixtures.NewMockRoundTripper())
	container.ContentWriter().SafelyWriteFile("/deps/some-pkg/gflowspkg.json", `{"files": []}`)
	container.ContentWriter().SafelyWriteFile("/deps/my-pkg/gflowspkg.json", `{"files": []}`)
	somePkg, _ := env.LoadDependency("/deps/some-pkg")
	myPkg, _ := env.LoadDependency("/deps/my-pkg")

	paths, err := env.GetLibPaths("my-workflow")
	assert.NoError(t, err)
	assert.Equal(t, []string{"/libs/some-lib", "/libs/my-lib", somePkg.LibsDir(), myPkg.LibsDir()}, paths)

	paths, err = env.GetLibPaths("other-workflow")
	assert.NoError(t, err)
	assert.Equal(t, []string{"/libs/some-lib", somePkg.LibsDir()}, paths)
}

func TestGetLibPathsAddContextLibsDir(t *testing.T) {
	config := strings.Join([]string{
		"templates:",
		"  engine: jsonnet",
		"  defaults:",
		"    libs: [/libs/some-lib]",
		"    dependencies: [/deps/some-pkg]",
	}, "\n")
	env, container := newTestEnv(config, fixtures.NewMockRoundTripper())
	container.ContentWriter().SafelyWriteFile("/deps/some-pkg/gflowspkg.json", `{"files": []}`)
	somePkg, _ := env.LoadDependency("/deps/some-pkg")
	container.FileSystem().Mkdir(".gflows/libs", 0644)

	paths, err := env.GetLibPaths("my-workflow")
	assert.NoError(t, err)
	assert.Equal(t, []string{"/libs/some-lib", somePkg.LibsDir(), ".gflows/libs"}, paths)
}

func TestCleanUpEnv(t *testing.T) {
	// arrange
	config := strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"  defaults:",
		"    dependencies: [/deps/some-pkg]",
	}, "\n")
	env, container := newTestEnv(config, fixtures.NewMockRoundTripper())
	fs := container.FileSystem()
	container.ContentWriter().SafelyWriteFile("/deps/some-pkg/gflowspkg.json", `{"files": []}`)

	lib, err := env.LoadDependency("/deps/some-pkg")
	assert.NoError(t, err)

	exists, err := fs.Exists(lib.LocalDir)
	assert.NoError(t, err)
	assert.True(t, exists, "expected LocalDir to exist")

	// act
	env.CleanUp()

	// assert
	exists, err = fs.Exists(lib.LocalDir)
	assert.NoError(t, err)
	assert.False(t, exists, "expected LocalDir to have been removed")
}
