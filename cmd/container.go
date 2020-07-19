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
	*adapters.Container
	context           *config.GFlowsContext
	workflowManager   *workflows.WorkflowManager
	workflowValidator *workflows.WorkflowValidator
	watcher           *workflows.Watcher
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

func CreateContainer(cmd *cobra.Command) (*Container, error) {
	adaptersContainer := adapters.CreateContainer()
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
		Container:         adaptersContainer,
		context:           context,
		workflowManager:   workflowManager,
		workflowValidator: workflowValidator,
		watcher:           watcher,
	}, nil
}
