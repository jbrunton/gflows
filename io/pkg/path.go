package pkg

import (
	"net/url"
	gopath "path"
	"path/filepath"
	"strings"
)

// IsRemotePath - returns true if the path is a URL, false otherwise
func IsRemotePath(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// ParentPath - returns the parent directory of the given path, for either local or remote paths
func ParentPath(path string) (string, error) {
	if !IsRemotePath(path) {
		return filepath.Dir(path), nil
	}

	parentURL, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	parentURL.Path = gopath.Dir(parentURL.Path)
	return parentURL.String(), nil
}

func JoinRelativePath(rootPath string, relPath string) (string, error) {
	if !IsRemotePath(rootPath) {
		return filepath.Join(rootPath, relPath), nil
	}

	URL, err := url.Parse(rootPath)
	if err != nil {
		return "", err
	}
	URL.Path = gopath.Join(URL.Path, relPath)
	return URL.String(), nil
}
