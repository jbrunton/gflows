package engine

import (
	"strings"
	"testing"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/workflow"
	"github.com/jbrunton/gflows/yamlutil"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func newJsonnetTemplateEngine(config string) (*content.Container, *config.GFlowsContext, *JsonnetTemplateEngine) {
	if config == "" {
		config = "templates:\n  engine: jsonnet"
	}
	ioContainer, context, _ := fixtures.NewTestContext(config)
	container := content.NewContainer(ioContainer)
	templateEngine := NewJsonnetTemplateEngine(container.FileSystem(), container.Logger(), context, container.ContentWriter())
	return container, context, templateEngine
}

func TestGenerateJsonnetWorkflowDefinitions(t *testing.T) {
	container, _, templateEngine := newJsonnetTemplateEngine("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(fixtures.ExampleJsonnetTemplate), 0644)

	definitions, _ := templateEngine.GetWorkflowDefinitions()

	expectedContent := fixtures.ExampleWorkflow("test.jsonnet")
	expectedJson, _ := yamlutil.YamlToJson(expectedContent)
	expectedDefinition := workflow.Definition{
		Name:        "test",
		Source:      ".gflows/workflows/test.jsonnet",
		Destination: ".github/workflows/test.yml",
		Content:     expectedContent,
		Status:      workflow.ValidationResult{Valid: true},
		JSON:        expectedJson,
	}
	assert.Equal(t, []*workflow.Definition{&expectedDefinition}, definitions)
}

func TestGetJsonnetWorkflowSources(t *testing.T) {
	container, _, templateEngine := newJsonnetTemplateEngine("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(fixtures.ExampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/test.libsonnet", []byte(fixtures.ExampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/invalid.ext", []byte(fixtures.ExampleJsonnetTemplate), 0644)

	sources := templateEngine.GetWorkflowSources()
	templates := templateEngine.GetWorkflowTemplates()

	assert.Equal(t, []string{".gflows/workflows/test.jsonnet", ".gflows/workflows/test.libsonnet"}, sources)
	assert.Equal(t, []string{".gflows/workflows/test.jsonnet"}, templates)
}

func TestGetJsonnetWorkflowName(t *testing.T) {
	_, _, templateEngine := newJsonnetTemplateEngine("")
	assert.Equal(t, "my-workflow-1", templateEngine.getWorkflowName("/workflows", "/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", templateEngine.getWorkflowName("/workflows", "/workflows/workflows/my-workflow-2.jsonnet"))
}

func TestGetJPath(t *testing.T) {
	config := strings.Join([]string{
		"templates:",
		"  engine: jsonnet",
		"  defaults:",
		"    libs: [some-lib]",
		"  overrides:",
		"    my-workflow:",
		"      libs: [my-lib]",
	}, "\n")
	_, _, engine := newJsonnetTemplateEngine(config)

	assert.Equal(t, []string{".gflows/some-lib", ".gflows/my-lib"}, engine.getJPath("my-workflow"))
	assert.Equal(t, []string{".gflows/some-lib"}, engine.getJPath("other-workflow"))
}
