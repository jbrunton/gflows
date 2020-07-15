package lib

import (
	"fmt"
	"io"
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
	fs  *afero.Afero
	out io.Writer
}

func NewContentWriter(fs *afero.Afero, out io.Writer) *ContentWriter {
	return &ContentWriter{
		fs:  fs,
		out: out,
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
	fmt.Fprintf(writer.out, "%11v %s %s\n", "error", destination, message)
	printStatusErrors(writer.out, errors, true)
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
		fmt.Fprintf(writer.out, "%11v %s %s\n", action, destination, details)
	} else {
		fmt.Fprintf(writer.out, "%11v %s\n", action, destination)
	}
}

func (writer *ContentWriter) ApplyGenerator(context *GFlowsContext, generator workflowGenerator) error {
	sourceFs, err := statikFs.New()
	if err != nil {
		return err
	}

	for _, sourcePath := range generator.sources {
		file, err := sourceFs.Open(sourcePath)
		if err != nil {
			return err
		}
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		destinationPath := filepath.Join(context.Dir, sourcePath)
		if err != nil {
			return err
		}
		writer.UpdateFileContent(destinationPath, string(content), "")
	}
	return nil
}
