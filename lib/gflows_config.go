package lib

import (
	"fmt"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// GFlowsConfig - type of current gflows context
type GFlowsConfig struct {
	GithubDir string `yaml:"githubDir"`
	Defaults  gflowsWorkflowConfig
	Workflows map[string]*gflowsWorkflowConfig
}

type gflowsWorkflowConfig struct {
	Checks gflowsChecksConfig
}

type gflowsChecksConfig struct {
	Schema  gflowsSchemaCheckConfig
	Content gflowsContentCheckConfig
}

type gflowsSchemaCheckConfig struct {
	Enabled *bool
	URI     string `yaml:"uri"`
}

type gflowsContentCheckConfig struct {
	Enabled *bool
}

// GetContextConfig - finds and returns the GFlowsConfig
func GetContextConfig(fs *afero.Afero, path string) (*GFlowsConfig, error) {
	data, err := fs.ReadFile(path)

	if err != nil {
		fmt.Println("Warning: no config set:", err)
		//data = []byte{}
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
