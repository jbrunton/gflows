package engine

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jbrunton/gflows/env"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/workflow"
	"github.com/jbrunton/gflows/yamlutil"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func newJsonnetTemplateEngine(config string, roundTripper http.RoundTripper) (*content.Container, *config.GFlowsContext, *JsonnetTemplateEngine) {
	if config == "" {
		config = "templates:\n  engine: jsonnet"
	}
	ioContainer, context, _ := fixtures.NewTestContext(config)
	container := content.NewContainer(ioContainer, &http.Client{Transport: roundTripper})
	installer := env.NewGFlowsLibInstaller(container.FileSystem(), container.ContentReader(), container.ContentWriter(), container.Logger())
	env := env.NewGFlowsEnv(container.FileSystem(), installer, context, container.Logger())
	templateEngine := NewJsonnetTemplateEngine(container.FileSystem(), context, container.ContentWriter(), env)
	return container, context, templateEngine
}

func TestGetJsonnetWorkflowDefinitions(t *testing.T) {
	container, _, templateEngine := newJsonnetTemplateEngine("", fixtures.NewMockRoundTripper())
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

func TestGetJsonnetWorkflowDefinitionsWithLibs(t *testing.T) {
	container, _, templateEngine := newJsonnetTemplateEngine("", fixtures.NewMockRoundTripper())
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(fixtures.ExampleJsonnetTemplate), 0644)
	container.ContentWriter().SafelyWriteFile("/path/to/my-lib.gflowslib", `{"libs": ["workflows/lib-workflow.jsonnet"]}`)
	container.ContentWriter().SafelyWriteFile("/path/to/workflows/lib-workflow.jsonnet", `std.manifestYamlDoc({})`)
	templateEngine.env.LoadLib("/path/to/my-lib.gflowslib")

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

func TestSerializationError(t *testing.T) {
	container, _, templateEngine := newJsonnetTemplateEngine("", fixtures.NewMockRoundTripper())
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte("{}"), 0644)

	definitions, _ := templateEngine.GetWorkflowDefinitions()

	expectedError := strings.Join([]string{
		"RUNTIME ERROR: expected string result, got: object",
		"\tDuring manifestation\t",
		"You probably need to serialize the output to YAML. See https://github.com/jbrunton/gflows/wiki/Templates#serialization",
	}, "\n")
	expectedDefinition := workflow.Definition{
		Name:        "test",
		Source:      ".gflows/workflows/test.jsonnet",
		Destination: ".github/workflows/test.yml",
		Content:     "",
		Status: workflow.ValidationResult{
			Valid:  false,
			Errors: []string{expectedError},
		},
		JSON: nil,
	}
	assert.Equal(t, []*workflow.Definition{&expectedDefinition}, definitions)
}

func TestGetJsonnetWorkflowSources(t *testing.T) {
	container, _, templateEngine := newJsonnetTemplateEngine("", fixtures.NewMockRoundTripper())
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
	_, _, templateEngine := newJsonnetTemplateEngine("", fixtures.NewMockRoundTripper())
	assert.Equal(t, "my-workflow-1", templateEngine.getWorkflowName("/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", templateEngine.getWorkflowName("/workflows/workflows/my-workflow-2.jsonnet"))
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
	_, _, engine := newJsonnetTemplateEngine(config, fixtures.NewMockRoundTripper())

	jpath, _ := engine.getJPath("my-workflow")
	assert.Equal(t, []string{".gflows/some-lib", ".gflows/my-lib"}, jpath)

	jpath, _ = engine.getJPath("other-workflow")
	assert.Equal(t, []string{".gflows/some-lib"}, jpath)
}

func TestJPathErrors(t *testing.T) {
	roundTripper := fixtures.NewMockRoundTripper()
	roundTripper.StubStatusCode("https://example.com/my-lib.gflowslib", 500)
	config := strings.Join([]string{
		"templates:",
		"  engine: jsonnet",
		"  defaults:",
		"    libs: [https://example.com/my-lib.gflowslib]",
	}, "\n")
	_, _, engine := newJsonnetTemplateEngine(config, roundTripper)

	_, err := engine.getJPath("my-workflow")
	assert.EqualError(t, err, "Received status code 500 from https://example.com/my-lib.gflowslib")
}
