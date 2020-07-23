package action

import (
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/workflows"

	"github.com/jbrunton/gflows/config"
)

type Container struct {
	*content.Container
	context         *config.GFlowsContext
	workflowManager *WorkflowManager
}

func (container *Container) Context() *config.GFlowsContext {
	return container.context
}

func (container *Container) WorkflowManager() *WorkflowManager {
	if container.workflowManager == nil {
		templateEngine := CreateWorkflowEngine(container.FileSystem(), container.Logger(), container.Context(), container.ContentWriter())
		container.workflowManager = NewWorkflowManager(
			container.FileSystem(),
			container.Logger(),
			container.Styles(),
			container.WorkflowValidator(),
			container.Context(),
			container.ContentWriter(),
			templateEngine)
	}
	return container.workflowManager
}

func (container *Container) WorkflowValidator() *workflows.WorkflowValidator {
	return workflows.NewWorkflowValidator(container.FileSystem(), container.Context())
}

func (container *Container) Watcher() *Watcher {
	return NewWatcher(container.WorkflowManager(), container.Context())
}

func NewContainer(parentContainer *content.Container, context *config.GFlowsContext) *Container {
	return &Container{
		Container: parentContainer,
		context:   context,
	}
}
