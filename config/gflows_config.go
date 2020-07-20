package config

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// GFlowsConfig - type of current gflows context
type GFlowsConfig struct {
	GithubDir string `yaml:"githubDir"`
	Defaults  GFlowsWorkflowConfig
	Workflows map[string]*GFlowsWorkflowConfig
	Jsonnet   jsonnetConfig
}

type jsonnetConfig struct {
	JPath []string `yaml:"jpath"`
}

type GFlowsWorkflowConfig struct {
	Checks GFlowsChecksConfig
}

type GFlowsChecksConfig struct {
	Schema  GFlowsSchemaCheckConfig
	Content GFlowsContentCheckConfig
}

type GFlowsSchemaCheckConfig struct {
	Enabled *bool
	URI     string `yaml:"uri"`
}

type GFlowsContentCheckConfig struct {
	Enabled *bool
}

// GetContextConfig - finds and returns the GFlowsConfig
func GetContextConfig(fs *afero.Afero, path string) (*GFlowsConfig, error) {
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

// GFlowsService - type of current gflows context
type GFlowsService struct {
	Name    string
	Version string
}

func parseConfig(input []byte) (*GFlowsConfig, error) {
	config := GFlowsConfig{}
	err := yaml.Unmarshal(input, &config)
	if err != nil {
		panic(err)
	}

	if config.Defaults.Checks.Schema.URI == "" {
		config.Defaults.Checks.Schema.URI = "https://json.schemastore.org/github-workflow"
	}

	return &config, nil
}
