package workflows

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/jbrunton/gflows/fixtures"
)

func setupTemplates(fs *afero.Afero) {
	fs.WriteFile(".gflows/workflows/foo.jsonnet", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/foo.libsonnet", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/nested/foo.jsonnet", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/bar/config.yml", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/bar/values.yml", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/invalid.ext", []byte(""), 0644)
}

func TestGetWorkflowSources(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	jsonnetEngine := NewJsonnetTemplateEngine(container.FileSystem(), container.Logger(), context)
	yttEngine := NewYttTemplateEngine(container.FileSystem(), container.Logger(), context)
	setupTemplates(container.FileSystem())

	scenarios := []struct {
		engine          string
		expectedSources []string
	}{
		{
			engine:          "jsonnet",
			expectedSources: []string{".gflows/workflows/foo.jsonnet", ".gflows/workflows/foo.libsonnet", ".gflows/workflows/nested/foo.jsonnet"},
		},
		{
			engine:          "ytt",
			expectedSources: []string{".gflows/workflows/bar/config.yml", ".gflows/workflows/bar/values.yml"},
		},
	}

	for _, scenario := range scenarios {
		context.Config.Templates.Defaults.Engine = scenario.engine
		manager := NewTemplateManager(context, jsonnetEngine, yttEngine)
		sources := manager.GetWorkflowSources()
		assert.Equal(t, scenario.expectedSources, sources, "Failures for scenario %+v", scenario)
	}
}

func TestGetWorkflowTemplates(t *testing.T) {
	container, context, _ := fixtures.NewTestContext("")
	jsonnetEngine := NewJsonnetTemplateEngine(container.FileSystem(), container.Logger(), context)
	yttEngine := NewYttTemplateEngine(container.FileSystem(), container.Logger(), context)
	setupTemplates(container.FileSystem())

	scenarios := []struct {
		engine            string
		expectedTemplates []string
	}{
		{
			engine:            "jsonnet",
			expectedTemplates: []string{".gflows/workflows/foo.jsonnet", ".gflows/workflows/nested/foo.jsonnet"},
		},
		{
			engine:            "ytt",
			expectedTemplates: []string{".gflows/workflows/bar"},
		},
	}

	for _, scenario := range scenarios {
		context.Config.Templates.Defaults.Engine = scenario.engine
		manager := NewTemplateManager(context, jsonnetEngine, yttEngine)
		sources := manager.GetWorkflowTemplates()
		assert.Equal(t, scenario.expectedTemplates, sources, "Failures for scenario %+v", scenario)
	}
}
