package workflows

import (
	"testing"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/content"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func newJsonnetTemplateEngine() (*content.Container, *config.GFlowsContext, *JsonnetTemplateEngine) {
	adaptersContainer, context, _ := fixtures.NewTestContext("templates:\n  engine: jsonnet")
	container := content.NewContainer(adaptersContainer)
	templateEngine := NewJsonnetTemplateEngine(container.FileSystem(), container.Logger(), context, container.ContentWriter())
	return container, context, templateEngine
}

func TestGenerateJsonnetWorkflowDefinitions(t *testing.T) {
	container, _, templateEngine := newJsonnetTemplateEngine()
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleJsonnetTemplate), 0644)

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
	container, _, templateEngine := newJsonnetTemplateEngine()
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/test.libsonnet", []byte(exampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/invalid.ext", []byte(exampleJsonnetTemplate), 0644)

	sources := templateEngine.GetWorkflowSources()
	templates := templateEngine.GetWorkflowTemplates()

	assert.Equal(t, []string{".gflows/workflows/test.jsonnet", ".gflows/workflows/test.libsonnet"}, sources)
	assert.Equal(t, []string{".gflows/workflows/test.jsonnet"}, templates)
}

func TestGetJsonnetWorkflowName(t *testing.T) {
	_, _, templateEngine := newJsonnetTemplateEngine()
	assert.Equal(t, "my-workflow-1", templateEngine.getWorkflowName("/workflows", "/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", templateEngine.getWorkflowName("/workflows", "/workflows/workflows/my-workflow-2.jsonnet"))
}
