package action

import (
	"fmt"

	"github.com/jbrunton/gflows/workflow"
)

func newTestWorkflowDefinition(name string, content string) *workflow.Definition {
	json, _ := workflow.YamlToJson(content)
	return &workflow.Definition{
		Name:        name,
		Source:      fmt.Sprintf(".gflows/workflows/%s.jsonnet", name),
		Destination: fmt.Sprintf(".github/workflows/%s.yml", name),
		Content:     content,
		JSON:        json,
	}
}
