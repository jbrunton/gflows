package workflows

import (
	"testing"

	"github.com/jbrunton/gflows/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestGenerateYttWorkflowDefinitions(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test.jsonnet", []byte(exampleTemplate), 0644)
	templateManager := NewYttTemplateManager(fs, container.Logger(), context)

	definitions, _ := templateManager.GetWorkflowDefinitions()

	expectedDefinition := WorkflowDefinition{
		Name:        "test",
		Source:      ".gflows/workflows/test.jsonnet",
		Destination: ".github/workflows/test.yml",
		Content:     exampleWorkflow("test"),
		Status:      ValidationResult{Valid: true},
	}
	assert.Equal(t, []*WorkflowDefinition{&expectedDefinition}, definitions)
}

func TestGetYttWorkflowSources(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/my-workflow/config1.yml", []byte("config1"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/config2.yaml", []byte("config2"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/config3.txt", []byte("config3"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/invalid.ext", []byte("ignored"), 0644)
	fs.WriteFile(".gflows/workflows/invalid-dir.yml", []byte("ignored"), 0644)
	templateManager := NewYttTemplateManager(fs, container.Logger(), context)

	sources := templateManager.GetWorkflowSources()

	assert.Equal(t, []string{".gflows/workflows/my-workflow/config1.yml", ".gflows/workflows/my-workflow/config2.yaml", ".gflows/workflows/my-workflow/config3.txt"}, sources)
}

func TestGetYttWorkflowTemplates(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/my-workflow/config1.yml", []byte("config1"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/nested-dir/config2.yaml", []byte("config2"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/config3.txt", []byte("config3"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/invalid.ext", []byte("ignored"), 0644)
	fs.WriteFile(".gflows/workflows/invalid-dir.yml", []byte("ignored"), 0644)
	fs.WriteFile(".gflows/workflows/another-workflow/config.yml", []byte("config"), 0644)
	templateManager := NewYttTemplateManager(fs, container.Logger(), context)

	templates := templateManager.GetWorkflowTemplates()

	assert.Equal(t, []string{".gflows/workflows/another-workflow", ".gflows/workflows/my-workflow"}, templates)
}
