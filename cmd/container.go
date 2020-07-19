package cmd

import (
	"os"

	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/workflows"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/config"
	"github.com/spf13/cobra"
)

type Container struct {
	*content.Container
	context         *config.GFlowsContext
	workflowManager *workflows.WorkflowManager
}

func (container *Container) Context() *config.GFlowsContext {
	return container.context
}

func (container *Container) WorkflowManager() *workflows.WorkflowManager {
	if container.workflowManager == nil {
		templateManager := workflows.NewJsonnetTemplateManager(container.FileSystem(), container.Logger(), container.Context())
		container.workflowManager = workflows.NewWorkflowManager(
			container.FileSystem(),
			container.Logger(),
			container.WorkflowValidator(),
			container.Context(),
			container.ContentWriter(),
			templateManager)
	}
	return container.workflowManager
}

func (container *Container) WorkflowValidator() *workflows.WorkflowValidator {
	return workflows.NewWorkflowValidator(container.FileSystem(), container.Context())
}

func (container *Container) Watcher() *workflows.Watcher {
	return workflows.NewWatcher(container.WorkflowManager(), container.Context())
}

func CreateContainer(cmd *cobra.Command) (*Container, error) {
	parentContainer := content.CreateContainer()
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
	return &Container{
		Container:       parentContainer,
		context:         context,
		workflowManager: workflowManager,
	}, nil
}
