package env

import (
	"encoding/json"
)

type GFlowsLibManifest struct {
	// Libs - the list of files in the library. If the manifest is remote, this list is used to
	// download the files.
	Libs []string
}

func ParseManifest(content string) (*GFlowsLibManifest, error) {
	manifest := GFlowsLibManifest{}
	err := json.Unmarshal([]byte(content), &manifest)
	return &manifest, err
}
