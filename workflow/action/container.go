package action

import (
	"github.com/jbrunton/gflows/env"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/workflow"

	"github.com/jbrunton/gflows/config"
)

type Container struct {
	*content.Container
	context         *config.GFlowsContext
	env             *env.GFlowsEnv
	installer       *env.GFlowsLibInstaller
	workflowManager *WorkflowManager
}

func (container *Container) Context() *config.GFlowsContext {
	return container.context
}

func (container *Container) WorkflowManager() *WorkflowManager {
	if container.workflowManager == nil {
		templateEngine := CreateWorkflowEngine(
			container.FileSystem(),
			container.Context(),
			container.ContentWriter(),
			container.Environment())
		container.workflowManager = NewWorkflowManager(
			container.FileSystem(),
			container.Logger(),
			container.Styles(),
			container.Validator(),
			container.Context(),
			container.ContentWriter(),
			templateEngine)
	}
	return container.workflowManager
}

func (container *Container) Environment() *env.GFlowsEnv {
	if container.env == nil {
		container.env = env.NewGFlowsEnv(
			container.FileSystem(),
			container.Installer(),
			container.Context(),
			container.Logger(),
		)
	}
	return container.env
}

func (container *Container) Installer() *env.GFlowsLibInstaller {
	if container.installer == nil {
		container.installer = env.NewGFlowsLibInstaller(
			container.FileSystem(),
			container.ContentReader(),
			container.ContentWriter(),
			container.Logger(),
		)
	}
	return container.installer
}

func (container *Container) Validator() *workflow.Validator {
	return workflow.NewValidator(container.FileSystem(), container.Context())
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
