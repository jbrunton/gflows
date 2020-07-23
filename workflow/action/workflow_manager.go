package action

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
	"github.com/jbrunton/gflows/workflow"
	"github.com/jbrunton/gflows/workflow/engine"
	statikFs "github.com/rakyll/statik/fs"

	fdiff "github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/spf13/afero"
)

func CreateWorkflowEngine(fs *afero.Afero, logger *adapters.Logger, context *config.GFlowsContext, contentWriter *content.Writer) workflow.TemplateEngine {
	var templateEngine workflow.TemplateEngine
	switch engineName := context.Config.Templates.Engine; engineName {
	case "jsonnet":
		templateEngine = engine.NewJsonnetTemplateEngine(fs, logger, context, contentWriter)
	case "ytt":
		templateEngine = engine.NewYttTemplateEngine(fs, logger, context, contentWriter)
	default:
		panic(fmt.Errorf("Unexpected engine: %s", engineName))
	}
	return templateEngine
}

type WorkflowManager struct {
	fs            *afero.Afero
	logger        *adapters.Logger
	styles        *styles.Styles
	validator     *workflow.WorkflowValidator
	context       *config.GFlowsContext
	contentWriter *content.Writer
	workflow.TemplateEngine
}

func NewWorkflowManager(
	fs *afero.Afero,
	logger *adapters.Logger,
	styles *styles.Styles,
	validator *workflow.WorkflowValidator,
	context *config.GFlowsContext,
	contentWriter *content.Writer,
	templateEngine workflow.TemplateEngine,
) *WorkflowManager {
	return &WorkflowManager{
		fs:             fs,
		logger:         logger,
		styles:         styles,
		validator:      validator,
		context:        context,
		contentWriter:  contentWriter,
		TemplateEngine: templateEngine,
	}
}

func (manager *WorkflowManager) GetWorkflows() []workflow.GitHubWorkflow {
	files := []string{}
	files, err := afero.Glob(manager.fs, filepath.Join(manager.context.GitHubDir, "workflows/*.yml"))
	if err != nil {
		panic(err)
	}

	definitions, err := manager.GetWorkflowDefinitions()
	if err != nil {
		panic(err) // TODO: improve handling
	}

	var gitHubWorkflows []workflow.GitHubWorkflow

	for _, file := range files {
		workflow := workflow.GitHubWorkflow{Path: file}
		for _, definition := range definitions {
			if definition.Destination == file {
				workflow.Definition = definition
				break
			}
		}
		gitHubWorkflows = append(gitHubWorkflows, workflow)
	}

	return gitHubWorkflows
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
		manager.logger.Printf("Checking %s ... ", manager.styles.Bold(definition.Name))

		if !definition.Status.Valid {
			manager.logger.Println(manager.styles.StyleError("FAILED"))
			manager.logger.Println("  Error parsing template:")
			manager.logger.PrintStatusErrors(definition.Status.Errors, false)
			valid = false
			continue
		}

		schemaResult := manager.validator.ValidateSchema(definition)
		if !schemaResult.Valid {
			manager.logger.Println(manager.styles.StyleError("FAILED"))
			manager.logger.Println("  Schema validation failed:")
			manager.logger.PrintStatusErrors(schemaResult.Errors, false)
			valid = false
		}

		contentResult := manager.validator.ValidateContent(definition)
		if !contentResult.Valid {
			if schemaResult.Valid { // otherwise we'll duplicate the failure message
				manager.logger.Println(manager.styles.StyleError("FAILED"))
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
			manager.logger.Println(manager.styles.StyleOK("OK"))
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

func (manager *WorkflowManager) InitWorkflows() {
	generator := manager.WorkflowGenerator()
	writer := content.NewWriter(manager.fs, manager.logger)
	sourceFs, err := statikFs.New()
	if err != nil {
		panic(err)
	}
	err = writer.ApplyGenerator(sourceFs, manager.context, generator)
	if err != nil {
		panic(err)
	}
}
