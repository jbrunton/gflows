package ytt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type FileSource struct {
	fs   *afero.Afero
	path string
	dir  string
}

func NewFileSource(fs *afero.Afero, path, dir string) FileSource { return FileSource{fs, path, dir} }

func (s FileSource) Description() string { return fmt.Sprintf("file '%s'", s.path) }

func (s FileSource) RelativePath() (string, error) {
	if s.dir == "" {
		return filepath.Base(s.path), nil
	}

	cleanPath, err := filepath.Abs(filepath.Clean(s.path))
	if err != nil {
		return "", err
	}

	cleanDir, err := filepath.Abs(filepath.Clean(s.dir))
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(cleanPath, cleanDir) {
		result := strings.TrimPrefix(cleanPath, cleanDir)
		result = strings.TrimPrefix(result, string(os.PathSeparator))
		return result, nil
	}

	return "", fmt.Errorf("unknown relative path for %s", s.path)
}

func (s FileSource) Bytes() ([]byte, error) { return s.fs.ReadFile(s.path) }
