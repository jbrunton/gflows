package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/jbrunton/gflows/adapters"
	_ "github.com/jbrunton/gflows/statik"
	"github.com/jbrunton/gflows/yaml"
	statikFs "github.com/rakyll/statik/fs"
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

// LoadConfig - finds and returns the GFlowsConfig
func LoadConfig(fs *afero.Afero, logger *adapters.Logger, opts ContextOpts) (config *GFlowsConfig, err error) {
	exists, err := fs.Exists(opts.ConfigPath)
	if !exists {
		if opts.Engine == "" {
			err = errors.New("no gflows context found")
		} else {
			config = &GFlowsConfig{}
			config.Templates.Engine = opts.Engine
			config.Workflows.Defaults.Checks.Schema.URI = "https://json.schemastore.org/github-workflow"
		}
		return
	}

	data, err := fs.ReadFile(opts.ConfigPath)
	if err != nil {
		return
	}

	err = validateConfig(string(data), logger)
	if err != nil {
		return
	}

	config, err = parseConfig(data)
	return
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

func validateConfig(config string, logger *adapters.Logger) error {
	json, err := yaml.YamlToJson(config)
	if err != nil {
		return err
	}

	sourceFs, err := statikFs.New()
	if err != nil {
		panic(err)
	}
	schemaFile, err := sourceFs.Open("/config-schema.json")
	if err != nil {
		panic(err)
	}
	defer schemaFile.Close()
	configSchema, err := ioutil.ReadAll(schemaFile)
	schemaLoader := gojsonschema.NewStringLoader(string(configSchema))
	configLoader := gojsonschema.NewGoLoader(json)
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
			logger.Println("Schema error:", err)
		}
		return errors.New("invalid config")
	}
	return nil
}
