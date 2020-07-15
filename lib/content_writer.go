package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	statikFs "github.com/rakyll/statik/fs"
	"github.com/spf13/afero"
)

type workflowGenerator struct {
	name    string
	sources []string
}

type ContentWriter struct {
	fs *afero.Afero
}

func NewContentWriter(fs *afero.Afero) *ContentWriter {
	return &ContentWriter{
		fs: fs,
	}
}

func (writer *ContentWriter) SafelyWriteFile(destination string, content string) error {
	dir := filepath.Dir(destination)
	if _, err := writer.fs.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	err := writer.fs.WriteFile(destination, []byte(content), 0644)
	return err
}

// LogErrors - prints an error message for the given destination file, together with any additional
func (writer *ContentWriter) LogErrors(destination string, message string, errors []string) {
	fmt.Printf("%11v %s %s\n", "error", destination, message)
	printStatusErrors(errors, true)
}

func (writer *ContentWriter) UpdateFileContent(destination string, content string, details string) {
	var action string
	exists, _ := writer.fs.Exists(destination)
	if exists {
		actualContent, _ := writer.fs.ReadFile(destination)
		if string(actualContent) == content {
			action = "identical"
		} else {
			action = "update"
		}
	} else {
		action = "create"
	}
	err := writer.SafelyWriteFile(destination, content)
	if err != nil {
		panic(err)
	}
	if details != "" {
		fmt.Printf("%11v %s %s\n", action, destination, details)
	} else {
		fmt.Printf("%11v %s\n", action, destination)
	}
}

func (writer *ContentWriter) ApplyGenerator(context *GFlowsContext, generator workflowGenerator) {
	sourceFs, err := statikFs.New()
	if err != nil {
		panic(err) // TODO: return this
	}

	for _, sourcePath := range generator.sources {
		file, err := sourceFs.Open(sourcePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1) // TODO: return this
		}
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		destinationPath := filepath.Join(context.Dir, sourcePath)
		if err != nil {
			panic(err)
		}
		writer.UpdateFileContent(destinationPath, string(content), "")
	}
}