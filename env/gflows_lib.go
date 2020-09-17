package env

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/config"
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
	// local then this will simply be ManifestPath. If ManifestPath is remote, then this will be
	// a local temp directory.
	LocalDir string

	// Files - the list of files in the library. If the manifest is remote, this list is used to
	// download the files.
	Files []string

	fs         *afero.Afero
	downloader *content.Downloader
	context    *config.GFlowsContext
}

type GFlowsLibManifest struct {
	Files []string
}

var libs map[string]*GFlowsLib

func NewGFlowsLib(manifestPath string, fs *afero.Afero, downloader *content.Downloader, context *config.GFlowsContext) *GFlowsLib {
	manifestName := filepath.Base(manifestPath)
	return &GFlowsLib{
		ManifestPath: manifestPath,
		ManifestName: manifestName,
		downloader:   downloader,
		fs:           fs,
		context:      context,
	}
}

func (lib *GFlowsLib) isRemote() bool {
	return strings.HasPrefix(lib.ManifestPath, "http://") || strings.HasPrefix(lib.ManifestPath, "https://")
}

func (lib *GFlowsLib) CleanUp() {
	if lib.isRemote() {
		fmt.Println("Removing temp directory", lib.LocalDir)
		lib.fs.RemoveAll(lib.LocalDir)
	}
}

func (lib *GFlowsLib) Download() error {
	if !lib.isRemote() {
		// TODO: maybe move this out of the lib class
		lib.LocalDir = lib.context.ResolvePath(path.Dir(lib.ManifestPath))
		fmt.Printf("Using %s (%s)\n", lib.ManifestName, lib.LocalDir)
		return nil
	}

	fmt.Printf("Downloading %s from %s...\n", lib.ManifestName, lib.ManifestPath)
	tempDir, err := lib.fs.TempDir("", lib.ManifestName)
	if err != nil {
		return err
	}
	lib.LocalDir = tempDir

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

	if err == nil {
		fmt.Printf("Downloaded and unpacked %s\n", lib.ManifestName)
	}

	return err
}

func (lib *GFlowsLib) downloadManifest() (*GFlowsLibManifest, error) {
	localPath := filepath.Join(lib.LocalDir, lib.ManifestName)
	err := lib.downloader.DownloadFile(lib.ManifestPath, localPath)
	if err != nil {
		return nil, err
	}

	manifestContent, err := lib.fs.ReadFile(localPath)
	if err != nil {
		return nil, err
	}

	manifest := GFlowsLibManifest{}
	err = json.Unmarshal(manifestContent, &manifest)
	return &manifest, err
}

func (lib *GFlowsLib) downloadLibFiles(rootUrl *url.URL, manifest *GFlowsLibManifest) error {
	for _, relPath := range manifest.Files {
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

func PushGFlowsLib(fs *afero.Afero, downloader *content.Downloader, libUrl string, context *config.GFlowsContext) (string, error) {
	lib := libs[libUrl]
	if lib != nil {
		// already processed
		return lib.LocalDir, nil
	}

	lib = NewGFlowsLib(libUrl, fs, downloader, context)
	err := lib.Download()
	if err != nil {
		return "", err
	}

	libs[libUrl] = lib
	return lib.LocalDir, nil
}

func CleanUpLibs() {
	for _, lib := range libs {
		lib.CleanUp()
	}
}

func init() {
	libs = make(map[string]*GFlowsLib)
}
