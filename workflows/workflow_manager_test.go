package workflows

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jbrunton/gflows/content"
	"github.com/spf13/afero"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func newTestWorkflowManager() (*afero.Afero, *bytes.Buffer, *WorkflowManager) {
	container, context, out := fixtures.NewTestContext("")
	fs := container.FileSystem()
	logger := adapters.NewLogger(out)
	styles := container.Styles()
	validator := NewWorkflowValidator(fs, context)
	contentWriter := content.NewWriter(fs, logger)
	templateEngine := CreateWorkflowEngine(fs, logger, context, contentWriter)
	return fs, out, NewWorkflowManager(
		fs,
		logger,
		styles,
		validator,
		context,
		contentWriter,
		templateEngine,
	)
}

func TestGetUnimportedWorkflows(t *testing.T) {
	fs, _, workflowManager := newTestWorkflowManager()
	fs.WriteFile(".github/workflows/workflow.yml", []byte(exampleWorkflow("test.jsonnet")), 0644)

	workflows := workflowManager.GetWorkflows()

	assert.Equal(t, []GitHubWorkflow{GitHubWorkflow{path: ".github/workflows/workflow.yml"}}, workflows)
}

func TestGetImportedWorkflows(t *testing.T) {
	fs, _, workflowManager := newTestWorkflowManager()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleJsonnetTemplate), 0644)
	fs.WriteFile(".github/workflows/test.yml", []byte(exampleWorkflow("test.jsonnet")), 0644)

	workflows := workflowManager.GetWorkflows()

	expectedContent := exampleWorkflow("test.jsonnet")
	expectedJson, _ := YamlToJson(expectedContent)
	expectedWorflow := GitHubWorkflow{
		path: ".github/workflows/test.yml",
		definition: &WorkflowDefinition{
			Name:        "test",
			Source:      ".gflows/workflows/test.jsonnet",
			Destination: ".github/workflows/test.yml",
			Content:     expectedContent,
			Status:      ValidationResult{Valid: true},
			JSON:        expectedJson,
		},
	}
	assert.Equal(t, []GitHubWorkflow{expectedWorflow}, workflows)
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
				fixtures.NewFile(".gflows/workflows/test.jsonnet", invalidJsonnetTemplate),
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
				fixtures.NewFile(".gflows/workflows/test.jsonnet", exampleJsonnetTemplate),
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
				fixtures.NewFile(".gflows/workflows/test.jsonnet", exampleJsonnetTemplate),
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
				fixtures.NewFile(".gflows/workflows/test.jsonnet", exampleJsonnetTemplate),
				fixtures.NewFile(".github/workflows/test.yml", exampleWorkflow("test.jsonnet")),
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
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleJsonnetTemplate), 0644)
	fs.WriteFile(".gflows/workflows/test2.jsonnet", []byte(exampleJsonnetTemplate), 0644)
	fs.WriteFile(".github/workflows/test.yml", []byte("out of date workflow"), 0644)

	err := workflowManager.UpdateWorkflows()

	assert.NoError(t, err)
	assert.Equal(t, strings.Join([]string{
		"     update .github/workflows/test.yml (from .gflows/workflows/test.jsonnet)",
		"     create .github/workflows/test2.yml (from .gflows/workflows/test2.jsonnet)",
	}, "\n")+"\n", out.String())
	testContent, _ := fs.ReadFile(".github/workflows/test.yml")
	assert.Equal(t, exampleWorkflow("test.jsonnet"), string(testContent))
	test2Content, _ := fs.ReadFile(".github/workflows/test2.yml")
	assert.Equal(t, exampleWorkflow("test2.jsonnet"), string(test2Content))
}
