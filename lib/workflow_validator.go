package lib

import (
	"fmt"

	"github.com/spf13/afero"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

// WorkflowValidator - validates a workflow definition
type WorkflowValidator struct {
	fs     *afero.Afero
	schema *gojsonschema.Schema
}

// ValidationResult - validate result
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// NewWorkflowValidator - creates a new validator for the given filesystem
func NewWorkflowValidator(fs *afero.Afero) *WorkflowValidator {
	// TODO: once https://github.com/SchemaStore/schemastore/pull/1135 is merged, switch back to https://json.schemastore.org/github-workflow
	schemaLoader := gojsonschema.NewReferenceLoader("https://raw.githubusercontent.com/SchemaStore/schemastore/1523f724449ef1f48228f85cae2464e3b5922bf6/src/schemas/json/github-workflow.json")
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		panic(err)
	}
	return &WorkflowValidator{
		fs:     fs,
		schema: schema,
	}
}

// ValidateSchema - validates the template for the definition generates a valid workflow
func (validator *WorkflowValidator) ValidateSchema(definition *WorkflowDefinition) ValidationResult {
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
	result, err := validator.schema.Validate(loader)
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
	exists, err := validator.fs.Exists(definition.Destination)
	if err != nil {
		panic(err)
	}

	if !exists {
		reason := fmt.Sprintf("Workflow missing for %q (expected workflow at %s)", definition.Name, definition.Destination)
		return ValidationResult{
			Valid:  false,
			Errors: []string{reason},
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
			Valid:  false,
			Errors: []string{reason},
		}
	}

	return ValidationResult{
		Valid:  true,
		Errors: []string{},
	}
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
			return nil, formatInvalidKeyError(keyPrefix, key)
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
