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

func (engine *JsonnetTemplateEngine) GetWorkflowSources() []string {
	files := []string{}
	err := engine.fs.Walk(engine.context.WorkflowsDir, func(path string, f os.FileInfo, err error) error {
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

func (engine *JsonnetTemplateEngine) GetWorkflowTemplates() []string {
	sources := engine.GetWorkflowSources()
	var templates []string
	for _, source := range sources {
		if filepath.Ext(source) == ".jsonnet" {
			templates = append(templates, source)
		}
	}
	return templates
}

// GetWorkflowDefinitions - get workflow definitions for the given context
func (engine *JsonnetTemplateEngine) GetWorkflowDefinitions() ([]*workflow.Definition, error) {
	templates := engine.GetWorkflowTemplates()
	definitions := []*workflow.Definition{}
	for _, templatePath := range templates {
		workflowName := engine.getWorkflowName(engine.context.WorkflowsDir, templatePath)
		vm := engine.createVM(workflowName)
		input, err := engine.fs.ReadFile(templatePath)
		if err != nil {
			return []*workflow.Definition{}, err
		}

		destinationPath := filepath.Join(engine.context.GitHubDir, "workflows/", workflowName+".yml")
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

func (engine *JsonnetTemplateEngine) ImportWorkflow(wf *workflow.GitHubWorkflow) (string, error) {
	workflowContent, err := engine.fs.ReadFile(wf.Path)
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
	templatePath := filepath.Join(engine.context.WorkflowsDir, templateName+".jsonnet")
	engine.contentWriter.SafelyWriteFile(templatePath, templateContent)

	return templatePath, nil
}

func (engine *JsonnetTemplateEngine) WorkflowGenerator() content.WorkflowGenerator {
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

func (engine *JsonnetTemplateEngine) getWorkflowName(workflowsDir string, filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}

func (engine *JsonnetTemplateEngine) createVM(workflowName string) *gojsonnet.VM {
	vm := gojsonnet.MakeVM()
	vm.Importer(&gojsonnet.FileImporter{
		JPaths: engine.getJPath(workflowName),
	})
	vm.StringOutput = true
	return vm
}

func (engine *JsonnetTemplateEngine) getJPath(workflowName string) []string {
	jpaths := engine.context.Config.GetTemplateArrayProperty(workflowName, func(config *config.GFlowsTemplateConfig) []string {
		return config.Jsonnet.JPath
	})

	return engine.context.ResolvePaths(jpaths)
}
