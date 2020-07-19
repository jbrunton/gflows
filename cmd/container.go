package cmd

import (
	"os"

	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/workflows"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Container struct {
	fileSystem        *afero.Afero
	logger            *adapters.Logger
	context           *config.GFlowsContext
	workflowManager   *workflows.WorkflowManager
	workflowValidator *workflows.WorkflowValidator
	watcher           *workflows.Watcher
}

func (container *Container) FileSystem() *afero.Afero {
	return container.fileSystem
}

func (container *Container) Logger() *adapters.Logger {
	return container.logger
}

func (container *Container) Context() *config.GFlowsContext {
	return container.context
}

func (container *Container) WorkflowManager() *workflows.WorkflowManager {
	return container.workflowManager
}

func (container *Container) WorkflowValidator() *workflows.WorkflowValidator {
	return container.workflowValidator
}

func (container *Container) Watcher() *workflows.Watcher {
	return container.watcher
}

func BuildContainer(cmd *cobra.Command) (*Container, error) {
	fs := adapters.CreateOsFs()
	context, err := config.GetContext(fs, cmd)
	if err != nil {
		return nil, err
	}
	logger := adapters.NewLogger(os.Stdout)
	contentWriter := content.NewWriter(fs, logger)
	templateManager := workflows.NewJsonnetTemplateManager(fs, logger, context)
	workflowValidator := workflows.NewWorkflowValidator(fs, context)
	workflowManager := workflows.NewWorkflowManager(fs, logger, workflowValidator, context, contentWriter, templateManager)
	watcher := workflows.NewWatcher(workflowManager, context)
	return &Container{
		fileSystem:        fs,
		logger:            adapters.NewLogger(os.Stdout),
		context:           context,
		workflowManager:   workflowManager,
		workflowValidator: workflowValidator,
		watcher:           watcher,
	}, nil
}

// func BuildContainer(fs *afero.Afero, logger *adapters.Logger, context *config.GFlowsContext) *Container {
// 	return &Container{
// 		fileSystem: fs,
// 		logger:     logger,
// 		context:    context,
// 	}
// }
