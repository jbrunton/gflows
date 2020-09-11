package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/afero"
)

type GFlowsLib struct {
	ManifestUrl string
	Files       []string
	TempDir     string
	FileSystem  *afero.Afero
}

type GFlowsLibManifest struct {
	Files []string
}

var libs []*GFlowsLib

func NewGFlowsLib(manifestUrl string, fs *afero.Afero) *GFlowsLib {
	lib := &GFlowsLib{
		ManifestUrl: manifestUrl,
		FileSystem:  fs,
	}
	libs = append(libs, lib)
	return lib
}

func (lib *GFlowsLib) CleanUp() {
	if lib.TempDir != "" {
		fmt.Println("Removing temp directory", lib.TempDir)
		lib.FileSystem.RemoveAll(lib.TempDir)
	}
}

func (lib *GFlowsLib) Download() error {
	fmt.Printf("Downloading lib from %s...\n", lib.ManifestUrl)
	manifestFilename := filepath.Base(lib.ManifestUrl)
	tempDir, err := lib.FileSystem.TempDir("", manifestFilename)
	if err != nil {
		panic(err)
	}

	rootUrl, err := url.Parse(lib.ManifestUrl)
	if err != nil {
		return err
	}
	rootUrl.Path = path.Dir(rootUrl.Path)

	manifestPath := filepath.Join(tempDir, manifestFilename)
	err = DownloadFile(lib.ManifestUrl, manifestPath, lib.FileSystem)
	if err != nil {
		return err
	}

	manifestContent, err := lib.FileSystem.ReadFile(manifestPath)
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
		err = DownloadFile(url.String(), dest, lib.FileSystem)
		if err != nil {
			return err
		}
	}

	lib.TempDir = tempDir

	fmt.Printf("Downloaded and unpacked %s\n", manifestFilename)

	return nil
}

func DownloadFile(url string, path string, fs *afero.Afero) error {
	// Create the file
	dir := filepath.Dir(path)
	if _, err := fs.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = fs.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	out, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file")
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("  Downloaded", url)

	return nil
}

func CleanUpLibs() {
	for _, lib := range libs {
		lib.CleanUp()
	}
}
