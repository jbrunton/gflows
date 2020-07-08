package lib

import (
	"fmt"

	"github.com/spf13/afero"
)

type fileSource struct {
	source      string
	destination string
	content     string
}

type workflowGenerator struct {
	name    string
	sources []string
}

func updateFileContent(fs *afero.Afero, destination string, content string, details string) {
	var action string
	exists, _ := fs.Exists(destination)
	if exists {
		actualContent, _ := fs.ReadFile(destination)
		if string(actualContent) == content {
			action = "  identical"
		} else {
			action = "     update"
		}
	} else {
		action = "     create"
	}
	err := safelyWriteFile(fs, destination, content)
	if err != nil {
		panic(err)
	}
	if details != "" {
		fmt.Println(action, destination, details)
	} else {
		fmt.Println(action, destination)
	}
}
