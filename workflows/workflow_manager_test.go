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
	fs, context, out := fixtures.NewTestContext("")
	logger := adapters.NewLogger(out)
	validator := NewWorkflowValidator(fs, context)
	contentWriter := content.NewWriter(fs, logger)
	templateManager := NewJsonnetTemplateManager(fs, logger, context)
	return fs, out, NewWorkflowManager(
		fs,
		logger,
		validator,
		context,
		contentWriter,
		templateManager,
	)
}

func TestGetWorkflowName(t *testing.T) {
	assert.Equal(t, "my-workflow-1", getWorkflowName("/workflows", "/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", getWorkflowName("/workflows", "/workflows/workflows/my-workflow-2.jsonnet"))
}

func TestValidateWorkflows(t *testing.T) {
	fs, _, workflowManager := newTestWorkflowManager()

	// invalid template
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(invalidTemplate), 0644)
	err := workflowManager.ValidateWorkflows(false)
	assert.EqualError(t, err, "workflow validation failed")

	// valid template, missing workflow
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	err = workflowManager.ValidateWorkflows(false)
	assert.EqualError(t, err, "workflow validation failed")

	// valid template, out of date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte("incorrect content"), 0644)
	err = workflowManager.ValidateWorkflows(false)
	assert.EqualError(t, err, "workflow validation failed")

	// valid template, up to date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte(exampleWorkflow), 0644)
	err = workflowManager.ValidateWorkflows(false)
	assert.NoError(t, err)
}

func TestValidateWorkflowsOutput(t *testing.T) {
	fs, out, workflowManager := newTestWorkflowManager()

	// invalid template
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(invalidTemplate), 0644)
	workflowManager.ValidateWorkflows(false)

	// valid template, missing workflow
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	workflowManager.ValidateWorkflows(false)

	// valid template, out of date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte("incorrect content"), 0644)
	workflowManager.ValidateWorkflows(false)

	// valid template, up to date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte(exampleWorkflow), 0644)
	workflowManager.ValidateWorkflows(false)

	expected := `
Checking [1mtest[0m ... [1;31mFAILED[0m
  Schema validation failed:
  ► (root): jobs is required
  Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
  ► Run "gflows workflow update" to update
Checking [1mtest[0m ... [1;31mFAILED[0m
  Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
  ► Run "gflows workflow update" to update
Checking [1mtest[0m ... [1;31mFAILED[0m
  Content is out of date for "test" (.github/workflows/test.yml)
  ► Run "gflows workflow update" to update
Checking [1mtest[0m ... [1;32mOK[0m
`
	assert.Equal(t, strings.TrimLeft(expected, "\n"), out.String())
}
