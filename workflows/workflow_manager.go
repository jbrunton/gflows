package workflows

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/di"
	"github.com/jbrunton/gflows/diff"
	"github.com/jbrunton/gflows/logs"
	"github.com/jbrunton/gflows/styles"
	"github.com/logrusorgru/aurora"

	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/google/go-jsonnet"
	statikFs "github.com/rakyll/statik/fs"
	"github.com/spf13/afero"
)

// WorkflowDefinition - definitoin for a workflow defined by a GFlows template
type WorkflowDefinition struct {
	Name        string
	Source      string
	Destination string
	Content     string
	Status      ValidationResult
}

type TemplateManager interface {
	// GetWorkflowSources - returns a list of all the files (i.e. templates + library files) used
	// to generate workflows.
	GetWorkflowSources() []string

	// GetWorkflowTemplates - returns a list of all the templates used to generator workflows.
	GetWorkflowTemplates() []string

	// GetWorkflowDefinitions - returns definitions generated from workflow templates.
	GetWorkflowDefinitions() ([]*WorkflowDefinition, error)
}

type WorkflowManager struct {
	fs            *afero.Afero
	logger        *logs.Logger
	validator     *WorkflowValidator
	context       *config.GFlowsContext
	contentWriter *content.Writer
	TemplateManager
}

type GitHubWorkflow struct {
	path       string
	definition *WorkflowDefinition
}

func NewWorkflowManager(container *di.Container) *WorkflowManager {
	return &WorkflowManager{
		fs:              container.FileSystem(),
		logger:          container.Logger(),
		validator:       NewWorkflowValidator(container),
		context:         container.Context(),
		contentWriter:   content.NewWriter(container),
		TemplateManager: NewJsonnetTemplateManager(container),
	}
}

func getWorkflowName(workflowsDir string, filename string) string {
	_, templateFileName := filepath.Split(filename)
	return strings.TrimSuffix(templateFileName, filepath.Ext(templateFileName))
}

func createVM(context *config.GFlowsContext) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: context.EvalJPaths(),
	})
	vm.StringOutput = true
	return vm
}

func (manager *WorkflowManager) GetWorkflows() []GitHubWorkflow {
	files := []string{}
	files, err := afero.Glob(manager.fs, filepath.Join(manager.context.GitHubDir, "workflows/*.yml"))
	if err != nil {
		panic(err)
	}

	definitions, err := manager.GetWorkflowDefinitions()
	if err != nil {
		panic(err) // TODO: improve handling
	}

	var workflows []GitHubWorkflow

	for _, file := range files {
		workflow := GitHubWorkflow{path: file}
		for _, definition := range definitions {
			if definition.Destination == file {
				workflow.definition = definition
				break
			}
		}
		workflows = append(workflows, workflow)
	}

	return workflows
}

// UpdateWorkflows - update workflow files for the given context
func (manager *WorkflowManager) UpdateWorkflows() error {
	definitions, err := manager.GetWorkflowDefinitions()
	if err != nil {
		return err
	}
	valid := true
	for _, definition := range definitions {
		details := fmt.Sprintf("(from %s)", definition.Source)
		if definition.Status.Valid {
			schemaResult := manager.validator.ValidateSchema(definition)
			if schemaResult.Valid {
				manager.contentWriter.UpdateFileContent(definition.Destination, definition.Content, details)
			} else {
				manager.contentWriter.LogErrors(definition.Destination, details, schemaResult.Errors)
				valid = false
			}
		} else {
			manager.contentWriter.LogErrors(definition.Destination, details, definition.Status.Errors)
			valid = false
		}
	}
	if !valid {
		return errors.New("errors encountered generating workflows")
	}
	return nil
}

// ValidateWorkflows - returns an error if the workflows are out of date
func (manager *WorkflowManager) ValidateWorkflows(showDiff bool) error {
	logger := logs.NewLogger(os.Stdout)
	definitions, err := manager.GetWorkflowDefinitions()
	if err != nil {
		return err
	}
	valid := true
	for _, definition := range definitions {
		fmt.Printf("Checking %s ... ", aurora.Bold(definition.Name))

		if !definition.Status.Valid {
			fmt.Println(styles.StyleError("FAILED"))
			fmt.Println("  Error parsing template:")
			logger.PrintStatusErrors(definition.Status.Errors, false)
			valid = false
			continue
		}

		schemaResult := manager.validator.ValidateSchema(definition)
		if !schemaResult.Valid {
			fmt.Println(styles.StyleError("FAILED"))
			fmt.Println("  Schema validation failed:")
			logger.PrintStatusErrors(schemaResult.Errors, false)
			valid = false
		}

		contentResult := manager.validator.ValidateContent(definition)
		if !contentResult.Valid {
			if schemaResult.Valid { // otherwise we'll duplicate the failure message
				fmt.Println(styles.StyleError("FAILED"))
			}
			fmt.Println("  " + contentResult.Errors[0])
			fmt.Println("  ► Run \"gflows workflow update\" to update")
			valid = false

			if showDiff {
				fpatch, err := diff.CreateFilePatch(contentResult.ActualContent, definition.Content)
				if err != nil {
					panic(err)
				}
				message := strings.Join([]string{
					fmt.Sprintf("src: <generated from: %s>\ndst: %s", definition.Source, definition.Destination),
					fmt.Sprintf(`This diff shows what will happen to %s if you run "gflows update"`, definition.Destination),
				}, "\n")
				patch := diff.NewPatch([]fdiff.FilePatch{fpatch}, message)
				logger.PrettyPrintDiff(patch.Format())
			}
		}

		if schemaResult.Valid && contentResult.Valid {
			fmt.Println(styles.StyleOK("OK"))
			for _, err := range append(schemaResult.Errors, contentResult.Errors...) {
				fmt.Printf("  Warning: %s\n", err)
			}
		}
	}
	if !valid {
		return errors.New("workflow validation failed")
	}
	return nil
}

// InitWorkflows - copies g3ops workflow sources to context directory
func InitWorkflows(container *di.Container) {
	generator := content.WorkflowGenerator{
		Name: "gflows",
		Sources: []string{
			"/workflows/common/steps.libsonnet",
			"/workflows/common/workflows.libsonnet",
			"/workflows/config/git.libsonnet",
			"/workflows/gflows.jsonnet",
			"/config.yml",
		},
	}
	writer := content.NewWriter(container)
	sourceFs, err := statikFs.New()
	if err != nil {
		err = writer.ApplyGenerator(sourceFs, container.Context(), generator)
	}
	if err != nil {
		fmt.Println(styles.StyleError(err.Error()))
	}
}
