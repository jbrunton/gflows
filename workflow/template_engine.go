package workflow

import (
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/pkg"
)

type TemplateEngine interface {
	// GetWorkflowSources - returns a list of all the files (i.e. templates + library files) used
	// to generate workflows.
	GetWorkflowSources() []string

	// GetWorkflowTemplates - returns a list of all the templates used to generate workflows.
	GetWorkflowTemplates() []*pkg.PathInfo

	// GetWorkflowDefinitions - returns definitions generated from workflow templates.
	GetWorkflowDefinitions() ([]*Definition, error)

	// ImportWorkflow - imports a workflow, returns the path to the new template.
	ImportWorkflow(workflow *GitHubWorkflow) (string, error)

	// WorkflowGenerator - returns a generator to create default workflow and config files
	WorkflowGenerator(templateVars map[string]string) content.WorkflowGenerator
}
