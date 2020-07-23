package workflow

import (
	"fmt"
)

func newTestWorkflowDefinition(name string, content string) *Definition {
	json, _ := YamlToJson(content)
	return &Definition{
		Name:        name,
		Source:      fmt.Sprintf(".gflows/workflows/%s.jsonnet", name),
		Destination: fmt.Sprintf(".github/workflows/%s.yml", name),
		Content:     content,
		JSON:        json,
	}
}
