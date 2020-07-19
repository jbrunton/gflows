package content

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/spf13/afero"
)

type WorkflowGenerator struct {
	Name    string
	Sources []string
}

type Writer struct {
	fs     *afero.Afero
	logger *adapters.Logger
}

func NewWriter(fs *afero.Afero, logger *adapters.Logger) *Writer {
	return &Writer{
		fs:     fs,
		logger: logger,
	}
}

func (writer *Writer) SafelyWriteFile(destination string, content string) error {
	dir := filepath.Dir(destination)
	if _, err := writer.fs.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = writer.fs.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	err := writer.fs.WriteFile(destination, []byte(content), 0644)
	return err
}

// LogErrors - prints an error message for the given destination file, together with any additional
func (writer *Writer) LogErrors(destination string, message string, errors []string) {
	writer.logger.Printfln("%11v %s %s", "error", destination, message)
	writer.logger.PrintStatusErrors(errors, true)
}

func (writer *Writer) UpdateFileContent(destination string, content string, details string) {
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
		writer.logger.Printfln("%11v %s %s", action, destination, details)
	} else {
		writer.logger.Printfln("%11v %s", action, destination)
	}
}

func (writer *Writer) ApplyGenerator(sourceFs http.FileSystem, context *config.GFlowsContext, generator WorkflowGenerator) error {
	for _, sourcePath := range generator.Sources {
		file, err := sourceFs.Open(sourcePath)
		if err != nil {
			return err
		}
		defer file.Close()
		content, err := ioutil.ReadAll(file)
		destinationPath := filepath.Join(context.Dir, strings.TrimPrefix(sourcePath, "/jsonnet"))
		if err != nil {
			return err
		}
		writer.UpdateFileContent(destinationPath, string(content), "")
	}
	return nil
}
