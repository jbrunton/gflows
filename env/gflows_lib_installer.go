package env

import (
	"fmt"
	"path"
	"strings"

	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/pkg"
	"github.com/spf13/afero"
)

type GFlowsLibInstaller struct {
	fs          *afero.Afero
	reader      *content.Reader
	writer      *content.Writer
	logger      *io.Logger
	repoManager *content.RepoManager
}

func NewGFlowsLibInstaller(
	fs *afero.Afero,
	reader *content.Reader,
	writer *content.Writer,
	logger *io.Logger,
	repoManager *content.RepoManager,
) *GFlowsLibInstaller {
	return &GFlowsLibInstaller{
		fs:          fs,
		reader:      reader,
		writer:      writer,
		logger:      logger,
		repoManager: repoManager,
	}
}

func (installer *GFlowsLibInstaller) install(lib *GFlowsLib) ([]*pkg.PathInfo, *GFlowsLibManifest, error) {
	if pkg.IsGitPath(lib.Path) {
		repoUrl, subdir := pkg.ParseGitPath(lib.Path)
		repo, err := installer.repoManager.GetRepo(repoUrl)
		if err != nil {
			return nil, nil, err
		}

		lib.ManifestPath, err = pkg.JoinRelativePath(repo.LocalDir, path.Join(subdir, "gflowspkg.json"))
		if err != nil {
			return nil, nil, err
		}
	}

	manifest, err := installer.loadManifest(lib.ManifestPath)
	if err != nil {
		return nil, nil, err
	}

	rootPath, err := pkg.ParentPath(lib.ManifestPath)
	if err != nil {
		return nil, nil, err
	}

	files := []*pkg.PathInfo{}
	for _, relPath := range manifest.Files {
		localPath, err := installer.copyFile(lib, rootPath, relPath)
		if err != nil {
			return nil, nil, err
		}

		pathInfo, err := lib.GetPathInfo(localPath)
		if err != nil {
			return nil, nil, err
		}

		files = append(files, pathInfo)
	}
	return files, manifest, nil
}

func (installer *GFlowsLibInstaller) loadManifest(manifestPath string) (*GFlowsLibManifest, error) {
	manifestContent, err := installer.reader.ReadContent(manifestPath)
	if err != nil {
		return nil, err
	}
	manifest, err := ParseManifest(manifestContent)
	if err == nil {
		if manifest.Libs != nil {
			installer.logger.Printfln(`WARNING: "libs" field is deprecated. Use "files" in %s`, manifestPath)
			manifest.Files = manifest.Libs
		}
	}
	return manifest, err
}

func (installer *GFlowsLibInstaller) copyFile(lib *GFlowsLib, rootPath string, relPath string) (string, error) {
	if !strings.HasPrefix(relPath, "libs/") && !strings.HasPrefix(relPath, "workflows/") {
		return "", fmt.Errorf("Unexpected directory %s, file must be in libs/ or workflows/", relPath)
	}
	sourcePath, err := pkg.JoinRelativePath(rootPath, relPath)
	if err != nil {
		return "", err
	}
	if pkg.IsRemotePath(rootPath) {
		installer.logger.Debugf("Downloading %s\n", sourcePath)
	} else {
		installer.logger.Debugf("Copying %s\n", sourcePath)
	}
	localPath, err := pkg.JoinRelativePath(lib.LocalDir, relPath)
	if err != nil {
		return "", err
	}
	sourceContent, err := installer.reader.ReadContent(sourcePath)
	if err != nil {
		return "", err
	}
	err = installer.writer.SafelyWriteFile(localPath, sourceContent)
	return localPath, err
}

func (installer *GFlowsLibInstaller) CleanUp() {
	installer.repoManager.CleanUp()
}
