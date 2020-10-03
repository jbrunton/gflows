package env

import (
	"encoding/json"
)

type GFlowsLibManifest struct {
	// Files - the list of files in the library. If the manifest is remote, this list is used to
	// download the files.
	Files []string

	// Libs - deprecated field, use Files instead
	Libs []string

	// Name - the name of the package
	Name string
}

func ParseManifest(content string) (*GFlowsLibManifest, error) {
	manifest := GFlowsLibManifest{}
	err := json.Unmarshal([]byte(content), &manifest)
	return &manifest, err
}
