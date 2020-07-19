package workflows

import (
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGenerateWorkflowDefinitions(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	templateManager := NewJsonnetTemplateManager(fs, container.Logger(), context)

	definitions, err := templateManager.GetWorkflowDefinitions()

	assert.NoError(t, err)
	assert.Len(t, definitions, 1)
	assert.Equal(t, ".gflows/workflows/test.jsonnet", definitions[0].Source)
	assert.Equal(t, ".github/workflows/test.yml", definitions[0].Destination)
	assert.Equal(t, definitions[0].Content, exampleWorkflow("test"))
}

func TestGetWorkflowName(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	templateManager := NewJsonnetTemplateManager(container.FileSystem(), container.Logger(), context)
	assert.Equal(t, "my-workflow-1", templateManager.getWorkflowName("/workflows", "/workflows/my-workflow-1.jsonnet"))
	assert.Equal(t, "my-workflow-2", templateManager.getWorkflowName("/workflows", "/workflows/workflows/my-workflow-2.jsonnet"))
}
