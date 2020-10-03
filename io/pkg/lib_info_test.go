package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/io"
)

func TestGetLibInfo(t *testing.T) {
	fs := io.CreateMemFs()
	fs.Mkdir("foo", 0644)
	fs.WriteFile("foo/bar.yml", []byte(""), 0644)

	scenarios := []struct {
		description  string
		libPath      string
		expectedInfo *LibInfo
	}{
		{
			description: "existing local file",
			libPath:     "foo/bar.yml",
			expectedInfo: &LibInfo{
				Path:     "foo/bar.yml",
				IsRemote: false,
				Exists:   true,
				IsDir:    false,
			},
		},
		{
			description: "non-existent local file",
			libPath:     "baz.yml",
			expectedInfo: &LibInfo{
				Path:     "baz.yml",
				IsRemote: false,
				Exists:   false,
				IsDir:    false,
			},
		},
		{
			description: "existing local dir",
			libPath:     "foo",
			expectedInfo: &LibInfo{
				Path:     "foo",
				IsRemote: false,
				Exists:   true,
				IsDir:    true,
			},
		},
		{
			description: "remote file",
			libPath:     "https://example.com/foo.yml",
			expectedInfo: &LibInfo{
				Path:     "https://example.com/foo.yml",
				IsRemote: true,
				Exists:   true,
				IsDir:    false,
			},
		},
	}

	for _, scenario := range scenarios {
		libInfo, err := GetLibInfo(scenario.libPath, fs)
		assert.NoError(t, err)
		assert.Equal(t, scenario.expectedInfo, libInfo, `Failures for scenario "%s"`, scenario.description)
	}
}
