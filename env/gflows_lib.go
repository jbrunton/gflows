package env

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/io/pkg"

	"github.com/davecgh/go-spew/spew"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/spf13/afero"
)

type GFlowsLib struct {
	// Path - the path to the package. Can be remote, local (relative) or local (absolute).
	Path string

	// ManifestPath - the path to the manifest, computed from Path.
	ManifestPath string

	// PackageName - the name of the package. This defaults to the name of the parent directory of
	// the manifest, but will be updated to the name in the manifest if given.
	PackageName string

	// LocalDir - the local directory of the library, to add to the lib paths. If ManifestPath is
	// local then this will simply be the directory containing ManifestPath. If ManifestPath is
	// remote, then this will be a local temp directory.
	LocalDir string

	// Files - content of the package as an array of FileInfo
	Files []*pkg.PathInfo

	fs        *afero.Afero
	installer *GFlowsLibInstaller
	context   *config.GFlowsContext
	logger    *io.Logger
}

func NewGFlowsLib(fs *afero.Afero, installer *GFlowsLibInstaller, logger *io.Logger, path string, context *config.GFlowsContext) (*GFlowsLib, error) {
	resolvedPath := context.ResolvePath(path)
	manifestPath, err := pkg.JoinRelativePath(resolvedPath, "gflowspkg.json")
	if err != nil {
		return nil, err
	}
	return &GFlowsLib{
		Path:         resolvedPath,
		ManifestPath: manifestPath,
		PackageName:  filepath.Base(resolvedPath),
		installer:    installer,
		fs:           fs,
		context:      context,
		logger:       logger,
	}, nil
}

func (lib *GFlowsLib) CleanUp() {
	lib.logger.Debug("Removing temp directory", lib.LocalDir)
	lib.fs.RemoveAll(lib.LocalDir)
}

func (lib *GFlowsLib) WorkflowsDir() string {
	return filepath.Join(lib.LocalDir, "/workflows")
}

func (lib *GFlowsLib) LibsDir() string {
	return filepath.Join(lib.LocalDir, "/libs")
}

func (lib *GFlowsLib) GetPathInfo(localPath string) (*pkg.PathInfo, error) {
	if !filepath.IsAbs(localPath) {
		return nil, fmt.Errorf("Expected %s to be absolute", localPath)
	}
	relPath, err := filepath.Rel(lib.LocalDir, localPath)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(relPath, "..") {
		return nil, fmt.Errorf("Expected %s to be a subdirectory of %s", localPath, lib.LocalDir)
	}
	rootPath, err := pkg.ParentPath(lib.ManifestPath)
	if err != nil {
		return nil, err
	}
	sourcePath, err := pkg.JoinRelativePath(rootPath, relPath)
	return &pkg.PathInfo{
		LocalPath:  localPath,
		SourcePath: sourcePath,
		// TODO: Description should be SourcePath if remote, relative SourcePath if within context dir (i.e. in source control),
		// and in terms of library name otherwise (since in that case the path is local but outside the repo, so not v useful)
		Description: path.Join(lib.PackageName, relPath),
	}, err
}

func (lib *GFlowsLib) Setup() error {
	lib.logger.Debugf("Installing %s (%s)\n", lib.PackageName, lib.Path)

	tempDir, err := lib.fs.TempDir("", lib.PackageName)
	if err != nil {
		return err
	}
	lib.LocalDir = tempDir

	files, manifest, err := lib.installer.install(lib)
	lib.Files = files

	if err == nil {
		lib.logger.Debugf("Installed %s\n", lib.PackageName)
		lib.logger.Debugf("Installed %s\n", spew.Sdump(lib.Files))

		if manifest.Name != "" {
			lib.PackageName = manifest.Name
		}
	}

	return err
}
