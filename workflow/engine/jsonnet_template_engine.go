package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/workflow"
	"github.com/jbrunton/gflows/yamlutil"

	gojsonnet "github.com/google/go-jsonnet"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/env"
	"github.com/jbrunton/gflows/workflow/engine/jsonnet"
	"github.com/spf13/afero"
)

type JsonnetTemplateEngine struct {
	fs            *afero.Afero
	context       *config.GFlowsContext
	contentWriter *content.Writer
	env           *env.GFlowsEnv
}

func NewJsonnetTemplateEngine(fs *afero.Afero, context *config.GFlowsContext, contentWriter *content.Writer, env *env.GFlowsEnv) *JsonnetTemplateEngine {
	return &JsonnetTemplateEngine{
		fs:            fs,
		context:       context,
		contentWriter: contentWriter,
		env:           env,
	}
}

func (engine *JsonnetTemplateEngine) GetWorkflowSources() []string {
	files := []string{}
	// TODO: replace GetWorkflowDirs() with GetLibs() and add .WorkflowDir() and .LibDir() to each lib.
	// ..altho: how to deal with local dir?
	// Perhaps add GFlowsPackage interface or something to abstract over libs and context.
	for _, workflowsDir := range engine.env.GetWorkflowDirs() {
		err := engine.fs.Walk(workflowsDir, func(path string, f os.FileInfo, err error) error {
			ext := filepath.Ext(path)
			if ext == ".jsonnet" || ext == ".libsonnet" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
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
		workflowName := engine.getWorkflowName(templatePath)
		vm, err := engine.createVM(workflowName)
		if err != nil {
			return []*workflow.Definition{}, err
		}
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
			errorDescription := strings.Trim(err.Error(), " \n\r")
			if strings.Contains(err.Error(), "expected string result") {
				errorDescription = strings.Join([]string{
					errorDescription,
					"You probably need to serialize the output to YAML. See https://github.com/jbrunton/gflows/wiki/Templates#serialization",
				}, "\n")
			}
			definition.Status.Errors = []string{errorDescription}
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

	normalizedContent, err := yamlutil.NormalizeWorkflow(string(workflowContent))
	if err != nil {
		return "", err
	}

	jsonData, err := yamlutil.YamlToJson(string(normalizedContent))
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
	templatePath := filepath.Join(engine.context.Dir, "workflows", templateName+".jsonnet")
	engine.contentWriter.SafelyWriteFile(templatePath, templateContent)

	return templatePath, nil
}

func (engine *JsonnetTemplateEngine) WorkflowGenerator(templateVars map[string]string) content.WorkflowGenerator {
	return content.WorkflowGenerator{
		Name:         "gflows",
		TemplateVars: templateVars,
		Sources: []content.WorkflowSource{
			content.NewWorkflowSource("/jsonnet/workflows/common/steps.libsonnet", "/workflows/common/steps.libsonnet"),
			content.NewWorkflowSource("/jsonnet/workflows/common/workflows.libsonnet", "/workflows/common/workflows.libsonnet"),
			content.NewWorkflowSource("/jsonnet/workflows/common/git.libsonnet", "/workflows/common/git.libsonnet"),
			content.NewWorkflowSource("/jsonnet/workflows/gflows.jsonnet", "/workflows/$WORKFLOW_NAME.jsonnet"),
			content.NewWorkflowSource("/jsonnet/config.yml", "/config.yml"),
		},
	}
}

func (engine *JsonnetTemplateEngine) getWorkflowName(filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}

func (engine *JsonnetTemplateEngine) createVM(workflowName string) (*gojsonnet.VM, error) {
	vm := gojsonnet.MakeVM()
	jpaths, err := engine.getJPath(workflowName)
	if err != nil {
		return nil, err
	}
	vm.Importer(&gojsonnet.FileImporter{
		JPaths: jpaths,
	})
	vm.StringOutput = true
	return vm, nil
}

func (engine *JsonnetTemplateEngine) getJPath(workflowName string) ([]string, error) {
	var jpaths []string
	for _, path := range engine.context.Config.GetTemplateLibs(workflowName) {
		if strings.HasSuffix(path, ".gflowslib") {
			lib, err := engine.env.LoadLib(path)
			if err != nil {
				return []string{}, err
			}
			cd, err := os.Getwd()
			if err != nil {
				return []string{}, err
			}
			libDir := lib.LocalDir
			if !filepath.IsAbs(libDir) {
				libDir = filepath.Join(cd, libDir)
			}
			jpaths = append(jpaths, libDir)
		} else {
			jpaths = append(jpaths, path)
		}
	}
	return engine.context.ResolvePaths(jpaths), nil
}
