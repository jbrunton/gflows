package workflows

import (
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGetWorkflowName(t *testing.T) {
	assert.Equal(t, "my-workflow-1", getWorkflowName("/workflows", "/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", getWorkflowName("/workflows", "/workflows/workflows/my-workflow-2.jsonnet"))
}

func TestGenerateWorkflowDefinitions(t *testing.T) {
	fs, context := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)

	definitions, err := GetWorkflowDefinitions(fs, context)

	assert.NoError(t, err)
	assert.Len(t, definitions, 1)
	assert.Equal(t, ".gflows/workflows/test.jsonnet", definitions[0].Source)
	assert.Equal(t, ".github/workflows/test.yml", definitions[0].Destination)
	assert.Equal(t, definitions[0].Content, exampleWorkflow)
}

func TestValidateWorkflows(t *testing.T) {
	fs, context := fixtures.NewTestContext(fixtures.NewTestCommand(), "")

	// invalid template
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(invalidTemplate), 0644)
	err := ValidateWorkflows(fs, context, false)
	assert.EqualError(t, err, "workflow validation failed")

	// valid template, missing workflow
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	err = ValidateWorkflows(fs, context, false)
	assert.EqualError(t, err, "workflow validation failed")

	// valid template, out of date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte("incorrect content"), 0644)
	err = ValidateWorkflows(fs, context, false)
	assert.EqualError(t, err, "workflow validation failed")

	// valid template, up to date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte(exampleWorkflow), 0644)
	err = ValidateWorkflows(fs, context, false)
	assert.NoError(t, err)
}

func ExampleValidateWorkflows() {
	fs, context := fixtures.NewTestContext(fixtures.NewTestCommand(), "")

	// invalid template
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(invalidTemplate), 0644)
	ValidateWorkflows(fs, context, false)

	// valid template, missing workflow
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	ValidateWorkflows(fs, context, false)

	// valid template, out of date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte("incorrect content"), 0644)
	ValidateWorkflows(fs, context, false)

	// valid template, up to date workflow
	fs.WriteFile(".github/workflows/test.yml", []byte(exampleWorkflow), 0644)
	ValidateWorkflows(fs, context, false)

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
