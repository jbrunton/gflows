package env

import (
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/spf13/afero"
)

type GFlowsLib struct {
	// ManifestPath - the path specified as the lib manifest. Can be remote, local (relative) or
	// local (absolute).
	ManifestPath string

	// ManifestName - the name of the manifest, which is the file name (without the path).
	// E.g. /path/to/my-manifest.gflowslib has the name `my-manifest.gflowslib`
	ManifestName string

	// LocalDir - the local directory of the library, to add to the lib paths. If ManifestPath is
	// local then this will simply be the directory containing ManifestPath. If ManifestPath is
	// remote, then this will be a local temp directory.
	LocalDir string

	// Files - content of the package as an array of FileInfo
	Files []*LibFileInfo

	fs        *afero.Afero
	installer *GFlowsLibInstaller
	context   *config.GFlowsContext
	logger    *io.Logger
}

func NewGFlowsLib(fs *afero.Afero, installer *GFlowsLibInstaller, logger *io.Logger, manifestPath string, context *config.GFlowsContext) *GFlowsLib {
	return &GFlowsLib{
		ManifestPath: context.ResolvePath(manifestPath),
		ManifestName: filepath.Base(manifestPath),
		installer:    installer,
		fs:           fs,
		context:      context,
		logger:       logger,
	}
}

func (lib *GFlowsLib) isRemote() bool {
	return strings.HasPrefix(lib.ManifestPath, "http://") || strings.HasPrefix(lib.ManifestPath, "https://")
}

func (lib *GFlowsLib) CleanUp() {
	lib.logger.Debug("Removing temp directory", lib.LocalDir)
	lib.fs.RemoveAll(lib.LocalDir)
}

func (lib *GFlowsLib) Setup() error {
	lib.logger.Debugf("Installing %s (%s)\n", lib.ManifestName, lib.ManifestPath)

	tempDir, err := lib.fs.TempDir("", lib.ManifestName)
	if err != nil {
		return err
	}
	lib.LocalDir = tempDir

	lib.Files, err = lib.installer.install(lib)

	if err == nil {
		lib.logger.Debugf("Installed %s\n", lib.ManifestName)
		lib.logger.Debugf("Installed %s\n", spew.Sdump(lib.Files))
	}

	return err
}
