package lib

import (
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func safelyWriteFile(fs *afero.Afero, destination string, content string) error {
	dir := filepath.Dir(destination)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	err := fs.WriteFile(destination, []byte(content), 0644)
	return err
}
