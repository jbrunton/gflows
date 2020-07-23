package workflow

import (
	"github.com/jbrunton/gflows/content"
)

type TemplateEngine interface {
	// GetWorkflowSources - returns a list of all the files (i.e. templates + library files) used
	// to generate workflow.
	GetWorkflowSources() []string

	// GetWorkflowTemplates - returns a list of all the templates used to generator workflow.
	GetWorkflowTemplates() []string

	// GetWorkflowDefinitions - returns definitions generated from workflow templates.
	GetWorkflowDefinitions() ([]*Definition, error)

	// ImportWorkflow - imports a workflow, returns the path to the new template.
	ImportWorkflow(workflow *GitHubWorkflow) (string, error)

	// WorkflowGenerator - returns a generator to create default workflow and config files
	WorkflowGenerator() content.WorkflowGenerator
}
