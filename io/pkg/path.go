package pkg

import (
	"net/url"
	gopath "path"
	"path/filepath"
	"regexp"
	"strings"
)

// IsGitPath - returns true if the path is for a Git repository, false otherwise
func IsGitPath(path string) bool {
	return strings.HasPrefix(path, "git@")
}

// ParseGitPath - returns the components of a Git path (the domain and the subdirectory)
func ParseGitPath(path string) (string, string) {
	r := regexp.MustCompile(`^(git@.*\.git)(.*)$`)
	matches := r.FindStringSubmatch(path)
	return matches[1], matches[2]
}

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
