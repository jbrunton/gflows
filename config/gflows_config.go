package config

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// GFlowsConfig - type of current gflows context
type GFlowsConfig struct {
	GithubDir string `yaml:"githubDir"`
	Workflows struct {
		Defaults  GFlowsWorkflowConfig
		Overrides map[string]*GFlowsWorkflowConfig
	}
	Templates struct {
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
	Engine  string
	Jsonnet struct {
		JPath []string `yaml:"jpath"`
	}
}

// LoadConfig - finds and returns the GFlowsConfig
func LoadConfig(fs *afero.Afero, path string) (*GFlowsConfig, error) {
	exists, err := fs.Exists(path)
	if !exists {
		return &GFlowsConfig{}, nil
	}

	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseConfig(data)
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
	var values []string
	workflowConfig := config.Templates.Overrides[workflowName]
	if workflowConfig != nil {
		values = selector(workflowConfig)
	}
	values = append(values, selector(&config.Templates.Defaults)...)
	return values
}

func parseConfig(input []byte) (*GFlowsConfig, error) {
	config := GFlowsConfig{}
	err := yaml.Unmarshal(input, &config)
	if err != nil {
		panic(err)
	}

	if config.Workflows.Defaults.Checks.Schema.URI == "" {
		config.Workflows.Defaults.Checks.Schema.URI = "https://json.schemastore.org/github-workflow"
	}

	return &config, nil
}
