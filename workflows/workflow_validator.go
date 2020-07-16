package workflows

import (
	"fmt"

	"github.com/jbrunton/gflows/config"
	"github.com/jbrunton/gflows/di"
	"github.com/spf13/afero"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

// WorkflowValidator - validates a workflow definition
type WorkflowValidator struct {
	fs            *afero.Afero
	defaultSchema *gojsonschema.Schema
	config        *config.GFlowsConfig
}

// ValidationResult - validate result
type ValidationResult struct {
	Valid         bool
	Errors        []string
	ActualContent string
}

// NewWorkflowValidator - creates a new validator for the given filesystem
func NewWorkflowValidator(container *di.Container) *WorkflowValidator {
	config := container.Context().Config
	schemaLoader := gojsonschema.NewReferenceLoader(config.Defaults.Checks.Schema.URI)
	defaultSchema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		panic(err)
	}
	return &WorkflowValidator{
		fs:            container.FileSystem(),
		defaultSchema: defaultSchema,
		config:        config,
	}
}

// ValidateSchema - validates the template for the definition generates a valid workflow
func (validator *WorkflowValidator) ValidateSchema(definition *WorkflowDefinition) ValidationResult {
	enabled := validator.getCheckEnabled(definition.Name, func(config config.GFlowsWorkflowConfig) *bool {
		return config.Checks.Schema.Enabled
	})
	if !enabled {
		return ValidationResult{
			Valid:  true,
			Errors: []string{fmt.Sprintf("Schema checks disabled for %s, skipping", definition.Name)},
		}
	}

	var yamlData map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(definition.Content), &yamlData)
	if err != nil {
		panic(err)
	}

	jsonData, err := convertToStringKeysRecursive(yamlData, "")
	if err != nil {
		panic(err)
	}

	loader := gojsonschema.NewGoLoader(jsonData)
	schema := validator.getWorkflowSchema(definition.Name)
	result, err := schema.Validate(loader)
	if err != nil {
		panic(err)
	}

	errors := []string{}
	for _, error := range result.Errors() {
		errors = append(errors, error.String())
	}

	return ValidationResult{
		Valid:  result.Valid(),
		Errors: errors,
	}
}

// ValidateContent - validates the content at the destination in the definition is up to date
func (validator *WorkflowValidator) ValidateContent(definition *WorkflowDefinition) ValidationResult {
	enabled := validator.getCheckEnabled(definition.Name, func(config config.GFlowsWorkflowConfig) *bool {
		return config.Checks.Content.Enabled
	})
	if !enabled {
		return ValidationResult{
			Valid:  true,
			Errors: []string{fmt.Sprintf("Content checks disabled for %s, skipping", definition.Name)},
		}
	}

	exists, err := validator.fs.Exists(definition.Destination)
	if err != nil {
		panic(err)
	}

	if !exists {
		reason := fmt.Sprintf("Workflow missing for %q (expected workflow at %s)", definition.Name, definition.Destination)
		return ValidationResult{
			Valid:         false,
			Errors:        []string{reason},
			ActualContent: "",
		}
	}

	data, err := validator.fs.ReadFile(definition.Destination)
	if err != nil {
		panic(err)
	}

	actualContent := string(data)
	if actualContent != definition.Content {
		reason := fmt.Sprintf("Content is out of date for %q (%s)", definition.Name, definition.Destination)
		return ValidationResult{
			Valid:         false,
			Errors:        []string{reason},
			ActualContent: actualContent,
		}
	}

	return ValidationResult{
		Valid:  true,
		Errors: []string{},
	}
}

func (validator *WorkflowValidator) getWorkflowSchema(workflowName string) *gojsonschema.Schema {
	workflowConfig := validator.config.Workflows[workflowName]
	if workflowConfig == nil || workflowConfig.Checks.Schema.URI == "" {
		return validator.defaultSchema
	}
	schemaLoader := gojsonschema.NewReferenceLoader(workflowConfig.Checks.Schema.URI)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		panic(err)
	}
	return schema
}

func (validator *WorkflowValidator) getCheckEnabled(workflowName string, selector func(config config.GFlowsWorkflowConfig) *bool) bool {
	workflowConfig := validator.config.Workflows[workflowName]
	if workflowConfig != nil {
		enabled := selector(*workflowConfig)
		if enabled != nil {
			return *enabled
		}
	}
	enabled := selector(validator.config.Defaults)
	if enabled != nil {
		return *enabled
	}
	return true
}

// Taken from Docker (and then refactored to keep CodeClimate happy).
// See: https://github.com/docker/docker-ce/blob/de14285fad39e215ea9763b8b404a37686811b3f/components/cli/cli/compose/loader/loader.go#L330
func convertToStringKeysRecursive(value interface{}, keyPrefix string) (interface{}, error) {
	if mapping, ok := value.(map[interface{}]interface{}); ok {
		return convertToStringDictKeysRecursive(mapping, keyPrefix)
	}
	if list, ok := value.([]interface{}); ok {
		return convertToStringListKeysRecursive(list, keyPrefix)
	}
	return value, nil
}

func convertToStringDictKeysRecursive(mapping map[interface{}]interface{}, keyPrefix string) (interface{}, error) {
	dict := make(map[string]interface{})
	for key, entry := range mapping {
		str, ok := key.(string)
		if !ok {
			if keyPrefix == "" && key == true {
				// Unfortunately GitHub workflows use the "on" reserved word, which canonically is treated as true, as a top
				// level key. We therefore guess that any key that gets parsed as true is intended to be used for the "on" key.
				str = "on"
			} else {
				return nil, formatInvalidKeyError(keyPrefix, key)
			}
		}
		var newKeyPrefix string
		if keyPrefix == "" {
			newKeyPrefix = str
		} else {
			newKeyPrefix = fmt.Sprintf("%s.%s", keyPrefix, str)
		}
		convertedEntry, err := convertToStringKeysRecursive(entry, newKeyPrefix)
		if err != nil {
			return nil, err
		}
		dict[str] = convertedEntry
	}
	return dict, nil
}

func convertToStringListKeysRecursive(list []interface{}, keyPrefix string) (interface{}, error) {
	var convertedList []interface{}
	for index, entry := range list {
		newKeyPrefix := fmt.Sprintf("%s[%d]", keyPrefix, index)
		convertedEntry, err := convertToStringKeysRecursive(entry, newKeyPrefix)
		if err != nil {
			return nil, err
		}
		convertedList = append(convertedList, convertedEntry)
	}
	return convertedList, nil
}

func formatInvalidKeyError(keyPrefix string, key interface{}) error {
	var location string
	if keyPrefix == "" {
		location = "at top level"
	} else {
		location = fmt.Sprintf("in %s", keyPrefix)
	}
	return fmt.Errorf("Non-string key %s: %#v", location, key)
}
