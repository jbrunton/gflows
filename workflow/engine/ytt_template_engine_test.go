package engine

import (
	"net/http"
	"strings"
	"testing"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/env"
	"github.com/jbrunton/gflows/fixtures"
	"github.com/jbrunton/gflows/io/content"
	"github.com/jbrunton/gflows/io/pkg"
	"github.com/jbrunton/gflows/workflow"
	"github.com/jbrunton/gflows/yamlutil"
	"github.com/stretchr/testify/assert"
)

func newYttTemplateEngine(config string) (*content.Container, *config.GFlowsContext, *YttTemplateEngine, *fixtures.MockRoundTripper) {
	ioContainer, context, _ := fixtures.NewTestContext(config)
	roundTripper := fixtures.NewMockRoundTripper()
	container := content.NewContainer(ioContainer, &http.Client{Transport: roundTripper})
	installer := env.NewGFlowsLibInstaller(container.FileSystem(), container.ContentReader(), container.ContentWriter(), container.Logger())
	env := env.NewGFlowsEnv(container.FileSystem(), installer, context, container.Logger())
	templateEngine := NewYttTemplateEngine(container.FileSystem(), context, container.ContentWriter(), env, container.Logger())
	return container, context, templateEngine, roundTripper
}

func TestGenerateYttWorkflowDefinitions(t *testing.T) {
	config := strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"  defaults:",
		"    dependencies:",
		"    - /my-pkg",
	}, "\n")
	container, _, templateEngine, _ := newYttTemplateEngine(config)
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/test/config.yml", []byte(""), 0644)
	fs.WriteFile("/my-pkg/gflowspkg.json", []byte(`{"files": []}`), 0644)

	definitions, _ := templateEngine.GetWorkflowDefinitions()

	expectedContent := "# File generated by gflows, do not modify\n# Source: .gflows/workflows/test\n"
	expectedJson, _ := yamlutil.YamlToJson(expectedContent)
	expectedDefinition := workflow.Definition{
		Name:        "test",
		Source:      ".gflows/workflows/test",
		Destination: ".github/workflows/test.yml",
		Description: ".gflows/workflows/test",
		Content:     expectedContent,
		Status:      workflow.ValidationResult{Valid: true},
		JSON:        expectedJson,
	}
	assert.Equal(t, []*workflow.Definition{&expectedDefinition}, definitions)
}

func TestGetYttObservableSources(t *testing.T) {
	config := strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"  defaults:",
		"    libs:",
		"    - vendor",
		"    - foo/bar.yml",
		"    - https://example.com/config.yml",
	}, "\n")
	container, _, templateEngine, _ := newYttTemplateEngine(config)
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/my-workflow/config1.yml", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/config2.yaml", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/config3.txt", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/invalid.ext", []byte(""), 0644)
	fs.WriteFile(".gflows/workflows/lib.yml", []byte(""), 0644)
	fs.WriteFile(".gflows/libs/lib.yml", []byte(""), 0644)
	fs.WriteFile("vendor/lib/config.yml", []byte(""), 0644)
	fs.WriteFile("foo/bar.yml", []byte(""), 0644)

	sources, err := templateEngine.GetObservableSources()

	assert.NoError(t, err)
	assert.Equal(t, []string{
		"vendor/lib/config.yml",
		"foo/bar.yml",
		".gflows/workflows/lib.yml",
		".gflows/workflows/my-workflow/config1.yml",
		".gflows/workflows/my-workflow/config2.yaml",
		".gflows/workflows/my-workflow/config3.txt",
		".gflows/libs/lib.yml",
	}, sources)
}

func TestGetYttWorkflowTemplates(t *testing.T) {
	container, _, templateEngine, _ := newYttTemplateEngine("")
	fs := container.FileSystem()
	fs.WriteFile(".gflows/workflows/my-workflow/config1.yml", []byte("config1"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/nested-dir/config2.yaml", []byte("config2"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/config3.txt", []byte("config3"), 0644)
	fs.WriteFile(".gflows/workflows/my-workflow/invalid.ext", []byte("ignored"), 0644)
	fs.WriteFile(".gflows/workflows/invalid-dir.yml", []byte("ignored"), 0644)
	fs.WriteFile(".gflows/workflows/another-workflow/config.yml", []byte("config"), 0644)
	fs.WriteFile(".gflows/workflows/jsonnet/foo.jsonnet", []byte("jsonnet"), 0644)

	templates, err := templateEngine.getWorkflowTemplates()

	expectedPaths := []*pkg.PathInfo{
		&pkg.PathInfo{
			SourcePath:  ".gflows/workflows/another-workflow",
			LocalPath:   ".gflows/workflows/another-workflow",
			Description: ".gflows/workflows/another-workflow",
		},
		&pkg.PathInfo{
			SourcePath:  ".gflows/workflows/my-workflow",
			LocalPath:   ".gflows/workflows/my-workflow",
			Description: ".gflows/workflows/my-workflow",
		},
	}
	assert.NoError(t, err)
	assert.Equal(t, expectedPaths, templates)
}

func TestGetAllYttLibs(t *testing.T) {
	config := strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"  defaults:",
		"    libs: [common, config]",
		"  overrides:",
		"    my-workflow:",
		"      libs: [my-lib]",
	}, "\n")
	_, _, engine, _ := newYttTemplateEngine(config)

	assert.Equal(t, []string{".gflows/common", ".gflows/config", ".gflows/my-lib"}, engine.getAllYttLibs())
}

func TestIsLib(t *testing.T) {
	config := strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"  defaults:",
		"    libs: [common, config]",
		"  overrides:",
		"    my-workflow:",
		"      libs: [my-lib]",
	}, "\n")
	_, _, engine, _ := newYttTemplateEngine(config)

	assert.Equal(t, true, engine.isLib(".gflows/common"))
	assert.Equal(t, true, engine.isLib(".gflows/my-lib"))
	assert.Equal(t, true, engine.isLib(".gflows/my-lib/"))
	assert.Equal(t, false, engine.isLib(".gflows/my-workflow.yml"))
}
