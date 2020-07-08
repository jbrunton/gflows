package lib

import (
	"path/filepath"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// JFlowsConfig - type of current jflows context
type JFlowsConfig struct {
	Name         string
	Environments map[string]jflowsEnvironmentConfig
	Services     map[string]jflowsServiceConfig
	Ci           jflowsCiConfig
	Workflows    jflowsWorkflowsConfig
	Repo         string
	Releases     jflowsReleasesConfig
}

type jflowsWorkflowsConfig struct {
	GitHubDir string `yaml:"githubDir"`
}

type jflowsEnvironmentConfig struct {
	Manifest string
}

type jflowsServiceConfig struct {
	Manifest string
}

type jflowsReleasesConfig struct {
	CreatePullRequest bool `yaml:"createPullRequest"`
}

type jflowsCiConfig struct {
	Defaults jflowsCiDefaultsConfig
}

type jflowsCiDefaultsConfig struct {
	Build jflowsBuildConfig
}

type jflowsBuildConfig struct {
	Env     map[string]string
	Command string
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

	for envName, env := range config.Environments {
		path, err := filepath.Abs(env.Manifest)
		if err != nil {
			panic(err)
		}
		env.Manifest = path
		config.Environments[envName] = env
	}

	for serviceName, service := range config.Services {
		path, err := filepath.Abs(service.Manifest)
		if err != nil {
			panic(err)
		}
		service.Manifest = path
		config.Services[serviceName] = service
	}

	return &config, nil
}
