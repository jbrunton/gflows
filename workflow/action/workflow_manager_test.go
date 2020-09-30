package action

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/jbrunton/gflows/env"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/workflow"
	"github.com/jbrunton/gflows/yamlutil"
	"github.com/spf13/afero"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func newTestWorkflowManager() (*afero.Afero, *bytes.Buffer, *WorkflowManager) {
	ioContainer, context, out := fixtures.NewTestContext("templates:\n  engine: jsonnet")
	httpClient := &http.Client{Transport: fixtures.NewMockRoundTripper()}
	fs := ioContainer.FileSystem()
	validator := workflow.NewValidator(fs, context)
	container := content.NewContainer(ioContainer, httpClient)
	installer := env.NewGFlowsLibInstaller(container.FileSystem(), container.ContentReader(), container.ContentWriter(), container.Logger())
	env := env.NewGFlowsEnv(fs, installer, context, container.Logger())
	templateEngine := CreateWorkflowEngine(fs, context, container.ContentWriter(), env)
	return fs, out, NewWorkflowManager(
		fs,
		container.Logger(),
		container.Styles(),
		validator,
		context,
		container.ContentWriter(),
		templateEngine,
	)
}

func TestGetUnimportedWorkflows(t *testing.T) {
	fs, _, workflowManager := newTestWorkflowManager()
	fs.WriteFile(".github/workflows/workflow.yml", []byte(fixtures.ExampleWorkflow("test.jsonnet")), 0644)

	gitHubWorkflows := workflowManager.GetWorkflows()

	assert.Equal(t, []workflow.GitHubWorkflow{workflow.GitHubWorkflow{Path: ".github/workflows/workflow.yml"}}, gitHubWorkflows)
}

func TestGetImportedWorkflows(t *testing.T) {
	fs, _, workflowManager := newTestWorkflowManager()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(fixtures.ExampleJsonnetTemplate), 0644)
	fs.WriteFile(".github/workflows/test.yml", []byte(fixtures.ExampleWorkflow("test.jsonnet")), 0644)

	gitHubWorkflows := workflowManager.GetWorkflows()

	expectedContent := fixtures.ExampleWorkflow("test.jsonnet")
	expectedJson, _ := yamlutil.YamlToJson(expectedContent)
	expectedWorflow := workflow.GitHubWorkflow{
		Path: ".github/workflows/test.yml",
		Definition: &workflow.Definition{
			Name:        "test",
			Source:      ".gflows/workflows/test.jsonnet",
			Description: ".gflows/workflows/test.jsonnet",
			Destination: ".github/workflows/test.yml",
			Content:     expectedContent,
			Status:      workflow.ValidationResult{Valid: true},
			JSON:        expectedJson,
		},
	}
	assert.Equal(t, []workflow.GitHubWorkflow{expectedWorflow}, gitHubWorkflows)
}

func TestValidateWorkflows(t *testing.T) {
	scenarios := []struct {
		description    string
		files          []fixtures.File
		expectedError  string
		expectedOutput string
	}{
		{
			description: "invalid template",
			files: []fixtures.File{
				fixtures.NewFile(".gflows/workflows/test.jsonnet", fixtures.InvalidJsonnetTemplate),
			},
			expectedError: "workflow validation failed",
			expectedOutput: `
Checking test ... FAILED
  Schema validation failed:
  ► (root): jobs is required
  Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
  ► Run "gflows workflow update" to update
`,
		},
		{
			description: "valid template, missing workflow",
			files: []fixtures.File{
				fixtures.NewFile(".gflows/workflows/test.jsonnet", fixtures.ExampleJsonnetTemplate),
			},
			expectedError: "workflow validation failed",
			expectedOutput: `
Checking test ... FAILED
  Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
  ► Run "gflows workflow update" to update
`,
		},
		{
			description: "valid template, out of date workflow",
			files: []fixtures.File{
				fixtures.NewFile(".gflows/workflows/test.jsonnet", fixtures.ExampleJsonnetTemplate),
				fixtures.NewFile(".github/workflows/test.yml", "incorrect content"),
			},
			expectedError: "workflow validation failed",
			expectedOutput: `
Checking test ... FAILED
  Content is out of date for "test" (.github/workflows/test.yml)
  ► Run "gflows workflow update" to update
`,
		},
		{
			description: "valid template, up to date workflow",
			files: []fixtures.File{
				fixtures.NewFile(".gflows/workflows/test.jsonnet", fixtures.ExampleJsonnetTemplate),
				fixtures.NewFile(".github/workflows/test.yml", fixtures.ExampleWorkflow("test.jsonnet")),
			},
			expectedOutput: `
Checking test ... OK
`,
		},
	}

	for _, scenario := range scenarios {
		fs, out, workflowManager := newTestWorkflowManager()
		for _, file := range scenario.files {
			file.Write(fs)
		}
		err := workflowManager.ValidateWorkflows(false)
		if scenario.expectedError == "" {
			assert.NoError(t, err, "Unexpected error for scenario %q", scenario.description)
		} else {
			assert.EqualError(t, err, scenario.expectedError, "Unexpected error for scenario %q", scenario.description)
		}
		assert.Equal(t, strings.TrimLeft(scenario.expectedOutput, "\n"), out.String(), "Failure for scenario %q", scenario.description)
	}

	// 	fs, out, workflowManager := newTestWorkflowManager()

	// 	// invalid template
	// 	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(invalidJsonnetTemplate), 0644)
	// 	workflowManager.ValidateWorkflows(false)

	// 	// valid template, missing workflow
	// 	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleJsonnetTemplate), 0644)
	// 	workflowManager.ValidateWorkflows(false)

	// 	// valid template, out of date workflow
	// 	fs.WriteFile(".github/workflows/test.yml", []byte("incorrect content"), 0644)
	// 	workflowManager.ValidateWorkflows(false)

	// 	// valid template, up to date workflow
	// 	fs.WriteFile(".github/workflows/test.yml", []byte(exampleWorkflow("test.jsonnet")), 0644)
	// 	workflowManager.ValidateWorkflows(false)

	// 	expected := `
	// Checking test ... FAILED
	//   Schema validation failed:
	//   ► (root): jobs is required
	//   Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
	//   ► Run "gflows workflow update" to update
	// Checking test ... FAILED
	//   Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
	//   ► Run "gflows workflow update" to update
	// Checking test ... FAILED
	//   Content is out of date for "test" (.github/workflows/test.yml)
	//   ► Run "gflows workflow update" to update
	// Checking test ... OK
	// `
	// 	fmt.Println("out:", out.String())
	// 	assert.Equal(t, strings.TrimLeft(expected, "\n"), out.String())
}

func TestUpdateWorkflows(t *testing.T) {
	fs, out, workflowManager := newTestWorkflowManager()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(fixtures.ExampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/test2.jsonnet", []byte(fixtures.ExampleJsonnetTemplate), 0644)
	fs.WriteFile(".github/workflows/test.yml", []byte("out of date workflow"), 0644)

	err := workflowManager.UpdateWorkflows()

	assert.NoError(t, err)
	assert.Equal(t, strings.Join([]string{
		"     update .github/workflows/test.yml (from .gflows/workflows/test.jsonnet)",
		"     create .github/workflows/test2.yml (from .gflows/workflows/test2.jsonnet)",
	}, "\n")+"\n", out.String())
	testContent, _ := fs.ReadFile(".github/workflows/test.yml")
	assert.Equal(t, fixtures.ExampleWorkflow("test.jsonnet"), string(testContent))
	test2Content, _ := fs.ReadFile(".github/workflows/test2.yml")
	assert.Equal(t, fixtures.ExampleWorkflow("test2.jsonnet"), string(test2Content))
}
