package workflow

import (
	"github.com/jbrunton/gflows/io/content"
)

type TemplateEngine interface {
	// GetObservableSources - returns a list of all the local files used to generate workflows. Used
	// to get the list of files to watch for changes.
	GetObservableSources() []string

	// GetWorkflowDefinitions - returns definitions generated from workflow templates.
	GetWorkflowDefinitions() ([]*Definition, error)

	// ImportWorkflow - imports a workflow, returns the path to the new template.
	ImportWorkflow(workflow *GitHubWorkflow) (string, error)

	// WorkflowGenerator - returns a generator to create default workflow and config files
	WorkflowGenerator(templateVars map[string]string) content.WorkflowGenerator
}
