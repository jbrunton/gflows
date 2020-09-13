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
	ManifestUrl string
	Files       []string
	TempDir     string
	fs          *afero.Afero
	downloader  *content.Downloader
}

type GFlowsLibManifest struct {
	Files []string
}

var libs map[string]*GFlowsLib

func NewGFlowsLib(manifestUrl string, fs *afero.Afero, downloader *content.Downloader) *GFlowsLib {
	return &GFlowsLib{
		ManifestUrl: manifestUrl,
		downloader:  downloader,
		fs:          fs,
	}
}

func (lib *GFlowsLib) CleanUp() {
	if lib.TempDir != "" {
		fmt.Println("Removing temp directory", lib.TempDir)
		lib.fs.RemoveAll(lib.TempDir)
	}
}

func (lib *GFlowsLib) Download() error {
	fmt.Printf("Downloading lib from %s...\n", lib.ManifestUrl)
	manifestFilename := filepath.Base(lib.ManifestUrl)
	tempDir, err := lib.fs.TempDir("", manifestFilename)
	if err != nil {
		panic(err)
	}

	rootUrl, err := url.Parse(lib.ManifestUrl)
	if err != nil {
		return err
	}
	rootUrl.Path = path.Dir(rootUrl.Path)

	manifestPath := filepath.Join(tempDir, manifestFilename)
	err = lib.downloader.DownloadFile(lib.ManifestUrl, manifestPath)
	if err != nil {
		return err
	}

	manifestContent, err := lib.fs.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	manifest := GFlowsLibManifest{}
	json.Unmarshal(manifestContent, &manifest)

	for _, relPath := range manifest.Files {
		// should be safe to ignore the error since we know it's valid
		url, _ := url.Parse(rootUrl.String())
		url.Path = path.Join(url.Path, relPath)
		dest := filepath.Join(tempDir, relPath)
		err = lib.downloader.DownloadFile(url.String(), dest)
		if err != nil {
			return err
		}
	}

	lib.TempDir = tempDir

	fmt.Printf("Downloaded and unpacked %s\n", manifestFilename)

	return nil
}

func PushGFlowsLib(fs *afero.Afero, downloader *content.Downloader, libUrl string) (string, error) {
	lib := libs[libUrl]
	if lib != nil {
		// already processed
		return lib.TempDir, nil
	}

	lib = NewGFlowsLib(libUrl, fs, downloader)
	lib.Download()
	libs[libUrl] = lib

	return lib.TempDir, nil
}

func CleanUpLibs() {
	for _, lib := range libs {
		lib.CleanUp()
	}
}

func init() {
	libs = make(map[string]*GFlowsLib)
}
