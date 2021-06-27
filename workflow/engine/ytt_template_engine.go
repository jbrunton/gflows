package engine

import (
	"os"
	"os/exec"
	"log"
	"bytes"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"github.com/jbrunton/gflows/io"
	"github.com/jbrunton/gflows/io/pkg"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/env"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/workflow"
	"github.com/jbrunton/gflows/yamlutil"
	"github.com/spf13/afero"
	"github.com/thoas/go-funk"
)

type YttTemplateEngine struct {
	fs            *afero.Afero
	context       *config.GFlowsContext
	contentWriter *content.Writer
	env           *env.GFlowsEnv
	logger        *io.Logger
}

func NewYttTemplateEngine(fs *afero.Afero, context *config.GFlowsContext, contentWriter *content.Writer, env *env.GFlowsEnv, logger *io.Logger) *YttTemplateEngine {
	return &YttTemplateEngine{
		fs:            fs,
		context:       context,
		contentWriter: contentWriter,
		env:           env,
		logger:        logger,
	}
}

func (engine *YttTemplateEngine) GetObservableSources() ([]string, error) {
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

		// If we reach here, it's a directory
		files = append(files, engine.getSourcesInDir(libPath)...)
	}
	return files, nil
}

func (engine *YttTemplateEngine) getWorkflowTemplates() ([]*pkg.PathInfo, error) {
	templates := []*pkg.PathInfo{}
	packages, err := engine.env.GetPackages()
	if err != nil {
		return nil, err
	}
	for _, pkg := range packages {
		paths, err := afero.Glob(engine.fs, filepath.Join(pkg.WorkflowsDir(), "/*"))
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			isDir, err := engine.fs.IsDir(path)
			if err != nil {
				return nil, err
			}
			if !isDir || engine.isLib(path) {
				continue
			}
			sources := engine.getSourcesInDir(path)
			if len(sources) > 0 {
				// only add directories with genuine source files
				pathInfo, err := pkg.GetPathInfo(path)
				if err != nil {
					return nil, err
				}
				templates = append(templates, pathInfo)
			}
		}
	}
	return templates, nil
}

// GetWorkflowDefinitions - get workflow definitions for the given context
func (engine *YttTemplateEngine) GetWorkflowDefinitions() ([]*workflow.Definition, error) {
	templates, err := engine.getWorkflowTemplates()
	if err != nil {
		return nil, err
	}
	definitions := []*workflow.Definition{}
	for _, template := range templates {
		workflowName := filepath.Base(template.LocalPath)
		destinationPath := filepath.Join(engine.context.GitHubDir, "workflows/", workflowName+".yml")
		definition := &workflow.Definition{
			Name:        workflowName,
			Source:      template.LocalPath,
			Description: template.Description,
			Destination: destinationPath,
			Status:      workflow.ValidationResult{Valid: true},
		}

		workflow, err := engine.apply(workflowName, template.LocalPath)

		if err != nil {
			definition.Status.Valid = false
			definition.Status.Errors = []string{strings.Trim(err.Error(), " \n\r")}
		} else {
			definition.SetContent(workflow, template)
		}

		definitions = append(definitions, definition)
	}

	return definitions, nil
}

func (engine *YttTemplateEngine) ImportWorkflow(workflow *workflow.GitHubWorkflow) (string, error) {
	workflowContent, err := engine.fs.ReadFile(workflow.Path)
	if err != nil {
		return "", err
	}

	templateContent, err := yamlutil.NormalizeWorkflow(string(workflowContent))
	if err != nil {
		return "", err
	}

	_, filename := filepath.Split(workflow.Path)
	templateName := strings.TrimSuffix(filename, filepath.Ext(filename))
	templatePath := filepath.Join(engine.context.WorkflowsDir(), templateName, templateName+".yml")
	engine.contentWriter.SafelyWriteFile(templatePath, string(templateContent))

	return templatePath, nil
}

func (engine *YttTemplateEngine) WorkflowGenerator(templateVars map[string]string) content.WorkflowGenerator {
	return content.WorkflowGenerator{
		Name:         "gflows",
		TemplateVars: templateVars,
		Sources: []content.WorkflowSource{
			content.NewWorkflowSource("/ytt/libs/steps.lib.yml", "/libs/steps.lib.yml"),
			content.NewWorkflowSource("/ytt/libs/workflows.lib.yml", "/libs/workflows.lib.yml"),
			content.NewWorkflowSource("/ytt/libs/values.yml", "/libs/values.yml"),
			content.NewWorkflowSource("/ytt/workflows/gflows/gflows.yml", "/workflows/$WORKFLOW_NAME/$WORKFLOW_NAME.yml"),
			content.NewWorkflowSource("/ytt/config.yml", "/config.yml"),
		},
	}
}

func (engine *YttTemplateEngine) getSourcesInDir(dir string) []string {
	files := []string{}
	err := engine.fs.Walk(dir, func(path string, f os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if ext == ".yml" || ext == ".yaml" || ext == ".txt" {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}

func (engine *YttTemplateEngine) getInputPaths(workflowName string, templateDir string) ([]string, error) {
	var sourcePaths []string
	for _, sourcePath := range engine.getSourcesInDir(templateDir) {
		sourcePaths = append(sourcePaths, sourcePath)
	}
	libPaths, _ := engine.env.GetLibPaths(workflowName)
	inputPaths := append(sourcePaths, libPaths...)
	engine.logger.Debugf("Input paths for %s: %s", workflowName, spew.Sdump(inputPaths))

	return inputPaths, nil
}

func (engine *YttTemplateEngine) apply(workflowName string, templateDir string) (string, error) {

	//cmd := exec.Command("bat", "--language", language, "--color", color, "--style", "plain")
	inputPaths, _ := engine.getInputPaths(workflowName, templateDir)
	var args []string
	for _, path := range inputPaths {
		args = append(args, "-f", path)
	}
	cmd := exec.Command("ytt", args...)
	//cmd.Stdin = strings.NewReader(code)


	var out bytes.Buffer
	cmd.Stdout = &out

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Print(stderr.String())
		panic(err);
		//return "", err
	}

	engine.logger.Println(out.String())

	return out.String(), nil
	// ui := cmdcore.NewPlainUI(false)
	// in, err := engine.getInput(workflowName, templateDir)
	// if err != nil {
	// 	return "", err
	// }
	// rootLibrary := workspace.NewRootLibrary(in.Files)

	// libraryExecutionFactory := workspace.NewLibraryExecutionFactory(ui, workspace.TemplateLoaderOpts{
	// 	IgnoreUnknownComments: true,
	// 	StrictYAML:            false,
	// })

	// libraryCtx := workspace.LibraryExecutionContext{Current: rootLibrary, Root: rootLibrary}
	// libraryLoader := libraryExecutionFactory.New(libraryCtx)

	// values, libraryValues, err := libraryLoader.Values([]*workspace.DataValues{}, &schema.AnySchema{})
	// if err != nil {
	// 	return "", err
	// }

	// result, err := libraryLoader.Eval(values, libraryValues)
	// if err != nil {
	// 	return "", err
	// }

	// workflowContent := ""

	// for _, file := range result.Files {
	// 	workflowContent = workflowContent + string(file.Bytes())
	// }

	// return workflowContent, nil
}

func (engine *YttTemplateEngine) getWorkflowName(workflowsDir string, filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}

func (engine *YttTemplateEngine) getAllYttLibs() []string {
	libs := engine.context.Config.Templates.Defaults.Libs

	for _, override := range engine.context.Config.Templates.Overrides {
		libs = append(libs, override.Libs...)
	}

	return engine.context.ResolvePaths(libs)
}

func (engine *YttTemplateEngine) isLib(path string) bool {
	_, isLib := funk.FindString(engine.getAllYttLibs(), func(lib string) bool {
		return filepath.Clean(lib) == filepath.Clean(path)
	})
	return isLib
}
