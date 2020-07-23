package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/workflow"

	gojsonnet "github.com/google/go-jsonnet"
	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/jsonnet"
	"github.com/spf13/afero"
)

type JsonnetTemplateEngine struct {
	fs            *afero.Afero
	logger        *adapters.Logger
	context       *config.GFlowsContext
	contentWriter *content.Writer
}

func NewJsonnetTemplateEngine(fs *afero.Afero, logger *adapters.Logger, context *config.GFlowsContext, contentWriter *content.Writer) *JsonnetTemplateEngine {
	return &JsonnetTemplateEngine{
		fs:            fs,
		logger:        logger,
		context:       context,
		contentWriter: contentWriter,
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
func (manager *JsonnetTemplateEngine) GetWorkflowDefinitions() ([]*workflow.Definition, error) {
	templates := manager.GetWorkflowTemplates()
	definitions := []*workflow.Definition{}
	for _, templatePath := range templates {
		workflowName := manager.getWorkflowName(manager.context.WorkflowsDir, templatePath)
		vm := manager.createVM(workflowName)
		input, err := manager.fs.ReadFile(templatePath)
		if err != nil {
			return []*workflow.Definition{}, err
		}

		destinationPath := filepath.Join(manager.context.GitHubDir, "workflows/", workflowName+".yml")
		definition := &workflow.Definition{
			Name:        workflowName,
			Source:      templatePath,
			Destination: destinationPath,
			Status:      workflow.ValidationResult{Valid: true},
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

func (manager *JsonnetTemplateEngine) ImportWorkflow(wf *workflow.GitHubWorkflow) (string, error) {
	workflowContent, err := manager.fs.ReadFile(wf.Path)
	if err != nil {
		return "", err
	}

	jsonData, err := workflow.YamlToJson(string(workflowContent))
	if err != nil {
		return "", err
	}

	json, err := jsonnet.Marshal(jsonData)
	if err != nil {
		return "", err
	}

	templateContent := fmt.Sprintf("local workflow = %s;\n\nstd.manifestYamlDoc(workflow)\n", string(json))

	_, filename := filepath.Split(wf.Path)
	templateName := strings.TrimSuffix(filename, filepath.Ext(filename))
	templatePath := filepath.Join(manager.context.WorkflowsDir, templateName+".jsonnet")
	manager.contentWriter.SafelyWriteFile(templatePath, templateContent)

	return templatePath, nil
}

func (manager *JsonnetTemplateEngine) WorkflowGenerator() content.WorkflowGenerator {
	return content.WorkflowGenerator{
		Name:       "gflows",
		TrimPrefix: "/jsonnet",
		Sources: []string{
			"/jsonnet/workflows/common/steps.libsonnet",
			"/jsonnet/workflows/common/workflows.libsonnet",
			"/jsonnet/workflows/config/git.libsonnet",
			"/jsonnet/workflows/gflows.jsonnet",
			"/jsonnet/config.yml",
		},
	}
}

func (manager *JsonnetTemplateEngine) getWorkflowName(workflowsDir string, filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}

func (manager *JsonnetTemplateEngine) createVM(workflowName string) *gojsonnet.VM {
	vm := gojsonnet.MakeVM()
	vm.Importer(&gojsonnet.FileImporter{
		JPaths: manager.evalJPaths(workflowName),
	})
	vm.StringOutput = true
	return vm
}

func (manager *JsonnetTemplateEngine) evalJPaths(workflowName string) []string {
	var paths []string
	configJPaths := manager.context.Config.GetTemplateArrayProperty(workflowName, func(config *config.GFlowsTemplateConfig) []string {
		return config.Jsonnet.JPath
	})

	for _, path := range configJPaths {
		if filepath.IsAbs(path) {
			paths = append(paths, path)
		} else {
			paths = append(paths, filepath.Join(manager.context.Dir, path))
		}
	}

	return paths
}
