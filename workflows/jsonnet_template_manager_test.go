package workflows

import (
	"testing"

	"github.com/jbrunton/gflows/adapters"
	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGenerateWorkflowDefinitions(t *testing.T) {
	fs, context, out := fixtures.NewTestContext(fixtures.NewTestCommand(), "")
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	templateManager := NewJsonnetTemplateManager(fs, adapters.NewLogger(out), context)

	definitions, err := templateManager.GetWorkflowDefinitions()

	assert.NoError(t, err)
	assert.Len(t, definitions, 1)
	assert.Equal(t, ".gflows/workflows/test.jsonnet", definitions[0].Source)
	assert.Equal(t, ".github/workflows/test.yml", definitions[0].Destination)
	assert.Equal(t, definitions[0].Content, exampleWorkflow)
}
