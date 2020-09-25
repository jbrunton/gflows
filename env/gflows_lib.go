package env

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/content"
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

	fs         *afero.Afero
	downloader *content.Downloader
	context    *config.GFlowsContext
	logger     *io.Logger
}

type GFlowsLibManifest struct {
	// Libs - the list of files in the library. If the manifest is remote, this list is used to
	// download the files.
	Libs []string
}

func NewGFlowsLib(fs *afero.Afero, downloader *content.Downloader, logger *io.Logger, manifestPath string, context *config.GFlowsContext) *GFlowsLib {
	manifestName := filepath.Base(manifestPath)
	return &GFlowsLib{
		ManifestPath: manifestPath,
		ManifestName: manifestName,
		downloader:   downloader,
		fs:           fs,
		context:      context,
		logger:       logger,
	}
}

func (lib *GFlowsLib) isRemote() bool {
	return strings.HasPrefix(lib.ManifestPath, "http://") || strings.HasPrefix(lib.ManifestPath, "https://")
}

func (lib *GFlowsLib) CleanUp() {
	//if lib.isRemote() {
	lib.logger.Debug("Removing temp directory", lib.LocalDir)
	lib.fs.RemoveAll(lib.LocalDir)
	//}
}

func (lib *GFlowsLib) Setup() error {
	lib.logger.Debugf("Installing %s (%s)\n", lib.ManifestName, lib.ManifestPath)

	tempDir, err := lib.fs.TempDir("", lib.ManifestName)
	if err != nil {
		return err
	}
	lib.LocalDir = tempDir

	if lib.isRemote() {
		err = lib.setupRemoteLib()
	} else {
		err = lib.setupLocalLib()
	}

	if err == nil {
		lib.logger.Debugf("Installed %s\n", lib.ManifestName)
	}

	return err
}

func (lib *GFlowsLib) setupLocalLib() error {
	manifestPath := lib.context.ResolvePath(lib.ManifestPath)
	manifestContent, err := lib.fs.ReadFile(manifestPath)
	lib.logger.Debugf("manifest: %s\n", string(manifestContent))
	if err != nil {
		return err
	}

	manifest := GFlowsLibManifest{}
	err = json.Unmarshal(manifestContent, &manifest)
	if err != nil {
		return err
	}

	rootPath := filepath.Dir(manifestPath)
	for _, relPath := range manifest.Libs {
		// should be safe to ignore the error since we know it's valid
		src := filepath.Join(rootPath, relPath)
		dest := filepath.Join(lib.LocalDir, relPath)
		fmt.Println("manifestPath:", manifestPath, "rootPath:", rootPath, "src:", src, "dest:", dest)
		err := lib.downloader.CopyFile(src, dest)
		if err != nil {
			return err
		}
	}
	return nil
}

func (lib *GFlowsLib) setupRemoteLib() error {
	rootUrl, err := url.Parse(lib.ManifestPath)
	if err != nil {
		return err
	}
	rootUrl.Path = path.Dir(rootUrl.Path)

	manifest, err := lib.downloadManifest()
	if err != nil {
		return err
	}

	err = lib.downloadLibFiles(rootUrl, manifest)

	return err
}

func (lib *GFlowsLib) downloadManifest() (*GFlowsLibManifest, error) {
	// TODO: could just read manifest, doesn't need to be stored on filesystem
	localPath := filepath.Join(lib.LocalDir, lib.ManifestName)
	err := lib.downloader.DownloadFile(lib.ManifestPath, localPath)
	if err != nil {
		return nil, err
	}

	manifestContent, err := lib.fs.ReadFile(localPath)
	lib.logger.Debugf("manifest: %s\n", string(manifestContent))
	if err != nil {
		return nil, err
	}

	manifest := GFlowsLibManifest{}
	err = json.Unmarshal(manifestContent, &manifest)
	return &manifest, err
}

func (lib *GFlowsLib) downloadLibFiles(rootUrl *url.URL, manifest *GFlowsLibManifest) error {
	for _, relPath := range manifest.Libs {
		// should be safe to ignore the error since we know it's valid
		url, _ := url.Parse(rootUrl.String())
		url.Path = path.Join(url.Path, relPath)
		dest := filepath.Join(lib.LocalDir, relPath)
		err := lib.downloader.DownloadFile(url.String(), dest)
		if err != nil {
			return err
		}
	}
	return nil
}
