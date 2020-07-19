package workflows

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/diff"
	"github.com/jbrunton/gflows/styles"
	"github.com/logrusorgru/aurora"

	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
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
	logger        *adapters.Logger
	validator     *WorkflowValidator
	context       *config.GFlowsContext
	contentWriter *content.Writer
	TemplateManager
}

type GitHubWorkflow struct {
	path       string
	definition *WorkflowDefinition
}

func NewWorkflowManager(
	fs *afero.Afero,
	logger *adapters.Logger,
	validator *WorkflowValidator,
	context *config.GFlowsContext,
	contentWriter *content.Writer,
	templateManager TemplateManager,
) *WorkflowManager {
	return &WorkflowManager{
		fs:              fs,
		logger:          logger,
		validator:       validator,
		context:         context,
		contentWriter:   contentWriter,
		TemplateManager: templateManager,
	}
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
	definitions, err := manager.GetWorkflowDefinitions()
	if err != nil {
		return err
	}
	valid := true
	for _, definition := range definitions {
		manager.logger.Printf("Checking %s ... ", aurora.Bold(definition.Name))

		if !definition.Status.Valid {
			manager.logger.Println(styles.StyleError("FAILED"))
			manager.logger.Println("  Error parsing template:")
			manager.logger.PrintStatusErrors(definition.Status.Errors, false)
			valid = false
			continue
		}

		schemaResult := manager.validator.ValidateSchema(definition)
		if !schemaResult.Valid {
			manager.logger.Println(styles.StyleError("FAILED"))
			manager.logger.Println("  Schema validation failed:")
			manager.logger.PrintStatusErrors(schemaResult.Errors, false)
			valid = false
		}

		contentResult := manager.validator.ValidateContent(definition)
		if !contentResult.Valid {
			if schemaResult.Valid { // otherwise we'll duplicate the failure message
				manager.logger.Println(styles.StyleError("FAILED"))
			}
			manager.logger.Println("  " + contentResult.Errors[0])
			manager.logger.Println("  ► Run \"gflows workflow update\" to update")
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
				manager.logger.PrettyPrintDiff(patch.Format())
			}
		}

		if schemaResult.Valid && contentResult.Valid {
			manager.logger.Println(styles.StyleOK("OK"))
			for _, err := range append(schemaResult.Errors, contentResult.Errors...) {
				manager.logger.Printf("  Warning: %s\n", err)
			}
		}
	}
	if !valid {
		return errors.New("workflow validation failed")
	}
	return nil
}

// InitWorkflows - copies g3ops workflow sources to context directory
func InitWorkflows(fs *afero.Afero, logger *adapters.Logger, context *config.GFlowsContext) {
	generator := content.WorkflowGenerator{
		Name: "gflows",
		Sources: []string{
			"/jsonnet/workflows/common/steps.libsonnet",
			"/jsonnet/workflows/common/workflows.libsonnet",
			"/jsonnet/workflows/config/git.libsonnet",
			"/jsonnet/workflows/gflows.jsonnet",
			"/config.yml",
		},
	}
	writer := content.NewWriter(fs, logger)
	sourceFs, err := statikFs.New()
	if err != nil {
		err = writer.ApplyGenerator(sourceFs, context, generator)
	}
	if err != nil {
		logger.Println(styles.StyleError(err.Error()))
	}
}
