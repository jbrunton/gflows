package workflows

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/spf13/afero"
)

type JsonnetTemplateEngine struct {
	fs      *afero.Afero
	logger  *adapters.Logger
	context *config.GFlowsContext
}

func NewJsonnetTemplateEngine(fs *afero.Afero, logger *adapters.Logger, context *config.GFlowsContext) *JsonnetTemplateEngine {
	return &JsonnetTemplateEngine{
		fs:      fs,
		logger:  logger,
		context: context,
	}
}

func (manager *JsonnetTemplateEngine) GetWorkflowSources() []string {
	files := []string{}
	err := manager.fs.Walk(manager.context.WorkflowsDir, func(path string, f os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext == ".jsonnet" || ext == ".libsonnet" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}

func (manager *JsonnetTemplateEngine) GetWorkflowTemplates() []string {
	sources := manager.GetWorkflowSources()
	var templates []string
	for _, source := range sources {
		if filepath.Ext(source) == ".jsonnet" {
			templates = append(templates, source)
		}
	}
	return templates
}

// GetWorkflowDefinitions - get workflow definitions for the given context
func (manager *JsonnetTemplateEngine) GetWorkflowDefinitions() ([]*WorkflowDefinition, error) {
	templates := manager.GetWorkflowTemplates()
	definitions := []*WorkflowDefinition{}
	for _, templatePath := range templates {
		workflowName := manager.getWorkflowName(manager.context.WorkflowsDir, templatePath)
		vm := createVM(manager.context, workflowName)
		input, err := manager.fs.ReadFile(templatePath)
		if err != nil {
			return []*WorkflowDefinition{}, err
		}

		destinationPath := filepath.Join(manager.context.GitHubDir, "workflows/", workflowName+".yml")
		definition := &WorkflowDefinition{
			Name:        workflowName,
			Source:      templatePath,
			Destination: destinationPath,
			Status:      ValidationResult{Valid: true},
		}

		workflow, err := vm.EvaluateSnippet(templatePath, string(input))

		if err != nil {
			definition.Status.Valid = false
			definition.Status.Errors = []string{strings.Trim(err.Error(), " \n\r")}
		} else {
			definition.SetContent(workflow, templatePath)
		}

		definitions = append(definitions, definition)
	}

	return definitions, nil
}

func (manager *JsonnetTemplateEngine) getWorkflowName(workflowsDir string, filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}

func createVM(context *config.GFlowsContext, workflowName string) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: context.EvalJPaths(workflowName),
	})
	vm.StringOutput = true
	return vm
}
