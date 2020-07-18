package workflows

import (
	"strings"
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGetWorkflowName(t *testing.T) {
	assert.Equal(t, "my-workflow-1", getWorkflowName("/workflows", "/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", getWorkflowName("/workflows", "/workflows/workflows/my-workflow-2.jsonnet"))
}

func TestValidateWorkflows(t *testing.T) {
	container, _ := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	fs := container.FileSystem()
	workflowManager := NewWorkflowManager(container)

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
	container, out := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	fs := container.FileSystem()
	workflowManager := NewWorkflowManager(container)

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

	assert.Equal(t, strings.Join([]string{
		`Checking [1mtest[0m ... [1;31mFAILED[0m`,
		`  Schema validation failed:`,
		`  ► (root): jobs is required`,
		`  Workflow missing for "test" (expected workflow at .github/workflows/test.yml)`,
		`  ► Run "gflows workflow update" to update`,
		`Checking [1mtest[0m ... [1;31mFAILED[0m`,
		`  Workflow missing for "test" (expected workflow at .github/workflows/test.yml)`,
		`  ► Run "gflows workflow update" to update`,
		`Checking [1mtest[0m ... [1;31mFAILED[0m`,
		`  Content is out of date for "test" (.github/workflows/test.yml)`,
		`  ► Run "gflows workflow update" to update`,
		`Checking [1mtest[0m ... [1;32mOK[0m`,
	}, "\n")+"\n", out.String())

	// Output:
	// Checking [1mtest[0m ... [1;31mFAILED[0m
	//   Schema validation failed:
	//   ► (root): jobs is required
	//   Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
	//   ► Run "gflows workflow update" to update
	// Checking [1mtest[0m ... [1;31mFAILED[0m
	//   Workflow missing for "test" (expected workflow at .github/workflows/test.yml)
	//   ► Run "gflows workflow update" to update
	// Checking [1mtest[0m ... [1;31mFAILED[0m
	//   Content is out of date for "test" (.github/workflows/test.yml)
	//   ► Run "gflows workflow update" to update
	// Checking [1mtest[0m ... [1;32mOK[0m
}
