package env

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"path/filepath"

	"github.com/jbrunton/gflows/io/content"
	"github.com/spf13/afero"
)

type GFlowsLib struct {
	ManifestUrl  string
	ManifestName string
	Files        []string
	tempDir      string
	fs           *afero.Afero
	downloader   *content.Downloader
}

type GFlowsLibManifest struct {
	Files []string
}

var libs map[string]*GFlowsLib

func NewGFlowsLib(manifestUrl string, fs *afero.Afero, downloader *content.Downloader) *GFlowsLib {
	manifestName := filepath.Base(manifestUrl)
	return &GFlowsLib{
		ManifestUrl:  manifestUrl,
		ManifestName: manifestName,
		downloader:   downloader,
		fs:           fs,
	}
}

func (lib *GFlowsLib) CleanUp() {
	if lib.tempDir != "" {
		fmt.Println("Removing temp directory", lib.tempDir)
		lib.fs.RemoveAll(lib.tempDir)
	}
}

func (lib *GFlowsLib) Download() error {
	fmt.Printf("Downloading %s from %s...\n", lib.ManifestName, lib.ManifestUrl)
	tempDir, err := lib.fs.TempDir("", lib.ManifestName)
	if err != nil {
		return err
	}
	lib.tempDir = tempDir

	rootUrl, err := url.Parse(lib.ManifestUrl)
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
	manifestPath := filepath.Join(lib.tempDir, lib.ManifestName)
	err := lib.downloader.DownloadFile(lib.ManifestUrl, manifestPath)
	if err != nil {
		return nil, err
	}

	manifestContent, err := lib.fs.ReadFile(manifestPath)
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
		dest := filepath.Join(lib.tempDir, relPath)
		err := lib.downloader.DownloadFile(url.String(), dest)
		if err != nil {
			return err
		}
	}
	return nil
}

func PushGFlowsLib(fs *afero.Afero, downloader *content.Downloader, libUrl string) (string, error) {
	lib := libs[libUrl]
	if lib != nil {
		// already processed
		return lib.tempDir, nil
	}

	lib = NewGFlowsLib(libUrl, fs, downloader)
	err := lib.Download()
	if err != nil {
		return "", err
	}

	libs[libUrl] = lib
	return lib.tempDir, nil
}

func CleanUpLibs() {
	for _, lib := range libs {
		lib.CleanUp()
	}
}

func init() {
	libs = make(map[string]*GFlowsLib)
}
