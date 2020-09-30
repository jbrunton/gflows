package fixtures

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func AssertTempDir(t *testing.T, fs *afero.Afero, dirName string, path string) {
	// On a Mac, temp dirs are in /var/folders. On GitHub, at /tmp.
	assert.Regexp(t, fmt.Sprintf("^(/var/folders|/tmp).*/%s[0-9]+$", dirName), path)
	exists, err := fs.Exists(path)
	assert.NoError(t, err)
	assert.True(t, exists, "expected %s to exist", path)
}
