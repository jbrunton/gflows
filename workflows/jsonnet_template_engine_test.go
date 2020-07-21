package workflows

import (
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJsonnetWorkflowDefinitions(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleJsonnetTemplate), 0644)
	templateEngine := NewJsonnetTemplateEngine(fs, container.Logger(), context)

	definitions, _ := templateEngine.GetWorkflowDefinitions()

	expectedContent := exampleWorkflow("test.jsonnet")
	expectedJson, _ := YamlToJson(expectedContent)
	expectedDefinition := WorkflowDefinition{
		Name:        "test",
		Source:      ".gflows/workflows/test.jsonnet",
		Destination: ".github/workflows/test.yml",
		Content:     expectedContent,
		Status:      ValidationResult{Valid: true},
		JSON:        expectedJson,
	}
	assert.Equal(t, []*WorkflowDefinition{&expectedDefinition}, definitions)
}

func TestGetJsonnetWorkflowSources(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/test.libsonnet", []byte(exampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/invalid.ext", []byte(exampleJsonnetTemplate), 0644)
	templateEngine := NewJsonnetTemplateEngine(fs, container.Logger(), context)

	sources := templateEngine.GetWorkflowSources()
	templates := templateEngine.GetWorkflowTemplates()

	assert.Equal(t, []string{".gflows/workflows/test.jsonnet", ".gflows/workflows/test.libsonnet"}, sources)
	assert.Equal(t, []string{".gflows/workflows/test.jsonnet"}, templates)
}

func TestGetJsonnetWorkflowName(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	templateEngine := NewJsonnetTemplateEngine(container.FileSystem(), container.Logger(), context)
	assert.Equal(t, "my-workflow-1", templateEngine.getWorkflowName("/workflows", "/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", templateEngine.getWorkflowName("/workflows", "/workflows/workflows/my-workflow-2.jsonnet"))
}
