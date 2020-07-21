package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/fixtures"
)

func TestGetWorkflowSources(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	jsonnetEngine := NewJsonnetTemplateEngine(container.FileSystem(), container.Logger(), context)
	yttEngine := NewYttTemplateEngine(container.FileSystem(), container.Logger(), context)

	container.FileSystem().WriteFile(".gflows/workflows/foo.jsonnet", []byte(""), 0644)
	container.FileSystem().WriteFile(".gflows/workflows/nested/foo.jsonnet", []byte(""), 0644)
	container.FileSystem().WriteFile(".gflows/workflows/bar/config.yml", []byte(""), 0644)
	container.FileSystem().WriteFile(".gflows/workflows/invalid.ext", []byte(""), 0644)

	scenarios := []struct {
		engine          string
		expectedSources []string
	}{
		{
			engine:          "jsonnet",
			expectedSources: []string{".gflows/workflows/foo.jsonnet", ".gflows/workflows/nested/foo.jsonnet"},
		},
		{
			engine:          "ytt",
			expectedSources: []string{".gflows/workflows/bar/config.yml"},
		},
	}

	for _, scenario := range scenarios {
		context.Config.Templates.Defaults.Engine = scenario.engine
		manager := NewTemplateManager(context, jsonnetEngine, yttEngine)
		sources := manager.GetWorkflowSources()
		assert.Equal(t, scenario.expectedSources, sources, "Failures for scenario %+v", scenario)
	}
}
