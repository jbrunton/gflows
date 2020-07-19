package workflows

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	cmdcore "github.com/k14s/ytt/pkg/cmd/core"
	cmdtpl "github.com/k14s/ytt/pkg/cmd/template"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/spf13/afero"
)

type YttTemplateManager struct {
	fs      *afero.Afero
	logger  *adapters.Logger
	context *config.GFlowsContext
}

func NewYttTemplateManager(fs *afero.Afero, logger *adapters.Logger, context *config.GFlowsContext) *YttTemplateManager {
	return &YttTemplateManager{
		fs:      fs,
		logger:  logger,
		context: context,
	}
}

func (manager *YttTemplateManager) getWorkflowSourcesInDir(dir string) []string {
	files := []string{}
	err := manager.fs.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Dir(path) == manager.context.WorkflowsDir {
			// ignore files in the top level workflows dir, as we need them to be in a nested directory to infer the template name
			return nil
		}
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

func (manager *YttTemplateManager) GetWorkflowSources() []string {
	return manager.getWorkflowSourcesInDir(manager.context.WorkflowsDir)
}

func (manager *YttTemplateManager) GetWorkflowTemplates() []string {
	// sources := manager.GetWorkflowSources()
	// var templates []string
	// for _, source := range sources {
	// 	relPath, _ := filepath.Rel(manager.context.WorkflowsDir, source)
	// 	topDir := filepath.SplitList(relPath)[0]
	// 	templates = append(templates, topDir)
	// }
	// return templates
	templates := []string{}
	paths, err := afero.Glob(manager.fs, filepath.Join(manager.context.WorkflowsDir, "/*"))
	if err != nil {
		panic(err)
	}
	for _, path := range paths {
		isDir, err := manager.fs.IsDir(path)
		if err != nil {
			panic(err)
		}
		if isDir {
			sources := manager.getWorkflowSourcesInDir(path)
			if len(sources) > 0 {
				// only add directories with genuine source files
				templates = append(templates, path)
			}
		}
	}
	return templates
}

type FileSource struct {
	fs   *afero.Afero
	path string
	dir  string
}

func NewFileSource(fs *afero.Afero, path, dir string) FileSource { return FileSource{fs, path, dir} }

func (s FileSource) Description() string { return fmt.Sprintf("file '%s'", s.path) }

func (s FileSource) RelativePath() (string, error) {
	if s.dir == "" {
		return filepath.Base(s.path), nil
	}

	cleanPath, err := filepath.Abs(filepath.Clean(s.path))
	if err != nil {
		return "", err
	}

	cleanDir, err := filepath.Abs(filepath.Clean(s.dir))
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(cleanPath, cleanDir) {
		result := strings.TrimPrefix(cleanPath, cleanDir)
		result = strings.TrimPrefix(result, string(os.PathSeparator))
		return result, nil
	}

	return "", fmt.Errorf("unknown relative path for %s", s.path)
}

func (s FileSource) Bytes() ([]byte, error) { return s.fs.ReadFile(s.path) }

func (manager *YttTemplateManager) getInput() cmdtpl.TemplateInput {
	var in cmdtpl.TemplateInput
	for _, sourcePath := range manager.GetWorkflowSources() {
		source := NewFileSource(manager.fs, sourcePath, filepath.Dir(sourcePath))
		file, err := files.NewFileFromSource(source)
		if err != nil {
			panic(err)
		}
		in.Files = append(in.Files, file)
	}
	return in
}

func (manager *YttTemplateManager) apply() (string, error) {
	ui := cmdcore.NewPlainUI(true)
	in := manager.getInput()
	rootLibrary := workspace.NewRootLibrary(in.Files)

	libraryExecutionFactory := workspace.NewLibraryExecutionFactory(ui, workspace.TemplateLoaderOpts{
		IgnoreUnknownComments: true,
		StrictYAML:            false,
	})

	libraryCtx := workspace.LibraryExecutionContext{Current: rootLibrary, Root: rootLibrary}
	libraryLoader := libraryExecutionFactory.New(libraryCtx)

	values, libraryValues, err := libraryLoader.Values([]*workspace.DataValues{})
	if err != nil {
		return "", err
	}
	// libraryValues = append(libraryValues, libraryValuesOverlays...)

	// if o.DataValuesFlags.Inspect {
	// 	return TemplateOutput{
	// 		DocSet: &yamlmeta.DocumentSet{
	// 			Items: []*yamlmeta.Document{values.Doc},
	// 		},
	// 	}
	// }

	result, err := libraryLoader.Eval(values, libraryValues)
	if err != nil {
		return "", err
	}

	workflowContent := ""

	for _, file := range result.Files {
		workflowContent = workflowContent + string(file.Bytes())
	}
	fmt.Println("content:", workflowContent)
	return workflowContent, nil
}

// GetWorkflowDefinitions - get workflow definitions for the given context
func (manager *YttTemplateManager) GetWorkflowDefinitions() ([]*WorkflowDefinition, error) {
	templates := manager.GetWorkflowTemplates()
	definitions := []*WorkflowDefinition{}
	for _, templatePath := range templates {
		// o := cmdtpl.NewOptions()
		// err := o.Run()
		// if err != nil {
		// 	panic(err)
		// }
		// vm := createVM(manager.context)
		// workflowName := manager.getWorkflowName(manager.context.WorkflowsDir, templatePath)
		// input, err := manager.fs.ReadFile(templatePath)
		// if err != nil {
		// 	return []*WorkflowDefinition{}, err
		// }

		workflowName := filepath.Base(templatePath)

		destinationPath := filepath.Join(manager.context.GitHubDir, "workflows/", workflowName+".yml")
		definition := &WorkflowDefinition{
			Name:        workflowName,
			Source:      templatePath,
			Destination: destinationPath,
			Status:      ValidationResult{Valid: true},
		}
		content, err := manager.apply()
		if err != nil {
			return nil, err
		}
		definition.Content = content

		// workflow, err := vm.EvaluateSnippet(templatePath, string(input))
		// if err != nil {
		// 	definition.Status.Valid = false
		// 	definition.Status.Errors = []string{strings.Trim(err.Error(), " \n\r")}
		// } else {
		// 	meta := strings.Join([]string{
		// 		"# File generated by gflows, do not modify",
		// 		fmt.Sprintf("# Source: %s", templatePath),
		// 	}, "\n")
		// 	definition.Content = meta + "\n" + workflow
		// }
		definitions = append(definitions, definition)
	}

	return definitions, nil
}

func (manager *YttTemplateManager) getWorkflowName(workflowsDir string, filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}
