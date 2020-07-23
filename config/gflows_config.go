package config

import (
	"errors"
	"fmt"

	"github.com/jbrunton/gflows/yaml"
	"github.com/spf13/afero"
	"github.com/thoas/go-funk"
	"github.com/xeipuuv/gojsonschema"
	goyaml "gopkg.in/yaml.v2"
)

// GFlowsConfig - type of current gflows context
type GFlowsConfig struct {
	GithubDir string `yaml:"githubDir"`
	Workflows struct {
		Defaults  GFlowsWorkflowConfig
		Overrides map[string]*GFlowsWorkflowConfig
	}
	Templates struct {
		Engine    string
		Defaults  GFlowsTemplateConfig
		Overrides map[string]*GFlowsTemplateConfig
	}
}

type GFlowsWorkflowConfig struct {
	Checks struct {
		Schema struct {
			Enabled *bool
			URI     string `yaml:"uri"`
		}
		Content struct {
			Enabled *bool
		}
	}
}

type GFlowsTemplateConfig struct {
	Libs []string
}

const configSchema = `
{
	"type": "object",
	"properties": {
		"githubDir": { "type": "string" }
	},
	"additionalProperties": false
}
`

// LoadConfig - finds and returns the GFlowsConfig
func LoadConfig(fs *afero.Afero, path string) (*GFlowsConfig, error) {
	exists, err := fs.Exists(path)
	if !exists {
		defaultConfig := &GFlowsConfig{}
		defaultConfig.Workflows.Defaults.Checks.Schema.URI = "https://json.schemastore.org/github-workflow"
		return defaultConfig, nil
	}

	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config, err := parseConfig(data)
	if err != nil {
		return nil, err
	}

	schemaLoader := gojsonschema.NewStringLoader(configSchema)
	configLoader := gojsonschema.NewStringLoader(yaml.CleanComments(string(data)))
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		panic(err)
	}
	result, err := schema.Validate(configLoader)
	if err != nil {
		panic(err)
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			fmt.Println("Schema error:", err)
		}
		return nil, errors.New("invalid config")
	}

	return config, nil
}

func (config *GFlowsConfig) GetWorkflowStringProperty(workflowName string, selector func(config *GFlowsWorkflowConfig) string) string {
	workflowConfig := config.Workflows.Overrides[workflowName]
	if workflowConfig != nil {
		value := selector(workflowConfig)
		if value != "" {
			return value
		}
	}
	return selector(&config.Workflows.Defaults)
}

func (config *GFlowsConfig) GetWorkflowBoolProperty(workflowName string, defaultValue bool, selector func(config *GFlowsWorkflowConfig) *bool) bool {
	workflowConfig := config.Workflows.Overrides[workflowName]
	if workflowConfig != nil {
		value := selector(workflowConfig)
		if value != nil {
			return *value
		}
	}
	value := selector(&config.Workflows.Defaults)
	if value != nil {
		return *value
	}
	return defaultValue
}

func (config *GFlowsConfig) GetTemplateArrayProperty(workflowName string, selector func(config *GFlowsTemplateConfig) []string) []string {
	values := selector(&config.Templates.Defaults)
	workflowConfig := config.Templates.Overrides[workflowName]
	if workflowConfig != nil {
		values = append(values, selector(workflowConfig)...)
	}
	return values
}

func (config *GFlowsConfig) GetTemplateLibs(workflowName string) []string {
	return config.GetTemplateArrayProperty(workflowName, func(config *GFlowsTemplateConfig) []string {
		return config.Libs
	})
}

func parseConfig(input []byte) (*GFlowsConfig, error) {
	config := GFlowsConfig{}
	err := goyaml.Unmarshal(input, &config)
	if err != nil {
		panic(err)
	}

	if config.Workflows.Defaults.Checks.Schema.URI == "" {
		config.Workflows.Defaults.Checks.Schema.URI = "https://json.schemastore.org/github-workflow"
	}
	if config.Templates.Engine == "" {
		return nil, errors.New("missing value for config: templates.engine")
	}
	if !funk.ContainsString([]string{"ytt", "jsonnet"}, config.Templates.Engine) {
		return nil, fmt.Errorf("unexpected value for templates.engine config field: %q (expected jsonnet or ytt)", config.Templates.Engine)
	}

	return &config, nil
}
