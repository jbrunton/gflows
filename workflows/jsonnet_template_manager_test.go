package workflows

import (
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGenerateWorkflowDefinitions(t *testing.T) {
	container, _ := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	templateManager := NewJsonnetTemplateManager(container)

	definitions, err := templateManager.GetWorkflowDefinitions()

	assert.NoError(t, err)
	assert.Len(t, definitions, 1)
	assert.Equal(t, ".gflows/workflows/test.jsonnet", definitions[0].Source)
	assert.Equal(t, ".github/workflows/test.yml", definitions[0].Destination)
	assert.Equal(t, definitions[0].Content, exampleWorkflow)
}
