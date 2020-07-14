package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	statikFs "github.com/rakyll/statik/fs"
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

func logUpdateError(destination string, details string, status ValidationResult) {
	fmt.Printf("%11v %s %s\n", "error", destination, details)
	printStatusErrors(status, true)
}

func updateFileContent(fs *afero.Afero, destination string, content string, details string) {
	var action string
	exists, _ := fs.Exists(destination)
	if exists {
		actualContent, _ := fs.ReadFile(destination)
		if string(actualContent) == content {
			action = "identical"
		} else {
			action = "update"
		}
	} else {
		action = "create"
	}
	err := safelyWriteFile(fs, destination, content)
	if err != nil {
		panic(err)
	}
	if details != "" {
		fmt.Printf("%11v %s %s\n", action, destination, details)
	} else {
		fmt.Printf("%11v %s\n", action, destination)
	}
}

func applyGenerator(fs *afero.Afero, context *GFlowsContext, generator workflowGenerator) {
	sourceFs, err := statikFs.New()
	if err != nil {
		panic(err)
	}

	for _, sourcePath := range generator.sources {
		file, err := sourceFs.Open(sourcePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		destinationPath := filepath.Join(context.Dir, sourcePath)
		if err != nil {
			panic(err)
		}
		source := fileSource{
			source:      sourcePath,
			destination: destinationPath,
			content:     string(content),
		}
		updateFileContent(fs, source.destination, source.content, "")
	}
}
