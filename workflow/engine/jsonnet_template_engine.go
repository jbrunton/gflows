package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/pkg"
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

func (engine *JsonnetTemplateEngine) GetObservableSources() ([]string, error) {
	files := []string{}
	for _, libPath := range append(
		engine.context.Config.GetAllLibs(),
		engine.context.WorkflowsDir(),
		engine.context.LibsDir(),
	) {
		libInfo, err := pkg.GetLibInfo(libPath, engine.fs)
		if err != nil {
			return nil, err
		}

		if libInfo.IsRemote || !libInfo.Exists {
			// Can't watch remote or non-existent files, so continue
			continue
		}

		// If it's a file...
		if !libInfo.IsDir {
			// ...add it to the list
			files = append(files, libPath)
			continue
		}

		// If we reach here, then libPath is a directory, so walk it
		err = engine.fs.Walk(libPath, func(path string, f os.FileInfo, err error) error {
			ext := filepath.Ext(path)
			// TODO: should probably include other files, since jsonnet can include json (and maybe text? any other types?)
			if ext == ".jsonnet" || ext == ".libsonnet" {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// GetWorkflowDefinitions - get workflow definitions for the given context
func (engine *JsonnetTemplateEngine) GetWorkflowDefinitions() ([]*workflow.Definition, error) {
	templates, err := engine.getWorkflowTemplates()
	if err != nil {
		return nil, err
	}
	definitions := []*workflow.Definition{}
	for _, template := range templates {
		workflowName := engine.getWorkflowName(template.LocalPath)
		vm, err := engine.createVM(workflowName)
		if err != nil {
			return []*workflow.Definition{}, err
		}
		input, err := engine.fs.ReadFile(template.LocalPath)
		if err != nil {
			return []*workflow.Definition{}, err
		}

		destinationPath := filepath.Join(engine.context.GitHubDir, "workflows/", workflowName+".yml")
		definition := &workflow.Definition{
			Name:        workflowName,
			Source:      template.LocalPath,
			Destination: destinationPath,
			Status:      workflow.ValidationResult{Valid: true},
		}

		workflow, err := vm.EvaluateSnippet(template.LocalPath, string(input))

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
			definition.SetContent(workflow, template)
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

func (engine *JsonnetTemplateEngine) getWorkflowTemplates() ([]*pkg.PathInfo, error) {
	templates := []*pkg.PathInfo{}
	packages, err := engine.env.GetPackages()
	if err != nil {
		return nil, err
	}
	for _, pkg := range packages {
		err := engine.fs.Walk(pkg.WorkflowsDir(), func(path string, f os.FileInfo, err error) error {
			ext := filepath.Ext(path)
			if ext == ".jsonnet" {
				pathInfo, err := pkg.GetPathInfo(path)
				if err != nil {
					return err
				}
				templates = append(templates, pathInfo)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return templates, nil
}

func (engine *JsonnetTemplateEngine) getWorkflowName(filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}

func (engine *JsonnetTemplateEngine) createVM(workflowName string) (*gojsonnet.VM, error) {
	vm := gojsonnet.MakeVM()
	jpaths, err := engine.env.GetLibPaths(workflowName)
	if err != nil {
		return nil, err
	}
	vm.Importer(&gojsonnet.FileImporter{
		JPaths: jpaths,
	})
	vm.StringOutput = true
	return vm, nil
}
