package content

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io"
	"github.com/spf13/afero"
)

type WorkflowSource struct {
	Source      string
	Destination string
}

func NewWorkflowSource(source string, destination string) WorkflowSource {
	return WorkflowSource{
		Source:      source,
		Destination: destination,
	}
}

type WorkflowGenerator struct {
	Name         string
	WorkflowName string
	Sources      []WorkflowSource
}

type Writer struct {
	fs     *afero.Afero
	logger *io.Logger
}

func NewWriter(fs *afero.Afero, logger *io.Logger) *Writer {
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
	for _, source := range generator.Sources {
		sourcePath := source.Source
		file, err := sourceFs.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("Error applying generator %s (file: %s)\n%s", generator.Name, sourcePath, err)
		}
		defer file.Close()
		destinationPath := filepath.Join(
			context.Dir,
			strings.ReplaceAll(source.Destination, "$WORKFLOW_NAME", generator.WorkflowName),
		)
		content, err := ioutil.ReadAll(file)
		if err != nil {
			return fmt.Errorf("Error applying generator %s (file: %s)\n%s", generator.Name, sourcePath, err)
		}
		writer.UpdateFileContent(destinationPath, string(content), "")
	}
	return nil
}
