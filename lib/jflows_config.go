package lib

import (
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// JFlowsConfig - type of current jflows context
type JFlowsConfig struct {
	GithubDir string `yaml:"githubDir"`
	Defaults  jflowsWorkflowConfig
	Workflows map[string]*jflowsWorkflowConfig
}

type jflowsWorkflowConfig struct {
	Checks jflowsChecksConfig
}

type jflowsChecksConfig struct {
	Schema  jflowsSchemaCheckConfig
	Content jflowsContentCheckConfig
}

type jflowsSchemaCheckConfig struct {
	Enabled *bool
	URI     string `yaml:"uri"`
}

type jflowsContentCheckConfig struct {
	Enabled *bool
}

// GetContextConfig - finds and returns the JFlowsConfig
func GetContextConfig(fs *afero.Afero, path string) (*JFlowsConfig, error) {
	data, err := fs.ReadFile(path)

	if err != nil {
		return nil, err
	}

	return parseConfig(data)
}

// JFlowsService - type of current jflows context
type JFlowsService struct {
	Name    string
	Version string
}

func parseConfig(input []byte) (*JFlowsConfig, error) {
	config := JFlowsConfig{}
	err := yaml.Unmarshal(input, &config)
	if err != nil {
		panic(err)
	}

	if config.Defaults.Checks.Schema.URI == "" {
		config.Defaults.Checks.Schema.URI = "https://json.schemastore.org/github-workflow"
	}

	return &config, nil
}
