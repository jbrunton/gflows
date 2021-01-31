package workflows

import (
	"github.com/jbrunton/gflows/content"
	"github.com/jbrunton/gflows/di"
	"github.com/spf13/cobra"
)

// type Container interface {
// 	FileSystem() *afero.Afero
// 	Logger() *logs.Logger
// 	Context() *config.GFlowsContext
// 	Validator() *WorkflowValidator
// 	ContentWriter() *content.Writer
// 	TemplateManager() TemplateManager
// }

type Container struct {
	*di.Container
}

func (container *Container) Validator() *WorkflowValidator {
	return NewWorkflowValidator(container.Container)
}

func (container *Container) ContentWriter() *content.Writer {
	return content.NewWriter(container.Container)
}

func (container *Container) TemplateManager() TemplateManager {
	return NewJsonnetTemplateManager(container.Container)
}

func (container *Container) WorkflowManager() *WorkflowManager {
	
}

func NewContainer(cmd *cobra.Command) (*Container, error) {
	parent, err := di.NewContainer(cmd)
	if err != nil {
		return nil, err
	}
	return &Container{parent}, nil
}
