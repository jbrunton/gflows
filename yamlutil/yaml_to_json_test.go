package yamlutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"encoding/json"
)

func TestYamlToJson(t *testing.T) {
	scenarios := []struct {
		description   string
		yaml          string
		expectedError string
		expectedJson  string
	}{
		{
			description:  "empty file",
			yaml:         "",
			expectedJson: "{}",
		},
		{
			description:  "simple yaml",
			yaml:         "foo: bar",
			expectedJson: `{"foo":"bar"}`,
		},
		{
			description:   "key error",
			yaml:          "123: foo",
			expectedError: "Non-string key at top level: 123",
		},
		{
			description:   "invalid yaml",
			yaml:          "&",
			expectedError: "yaml: did not find expected alphabetic or numeric character",
		},
	}

	for _, scenario := range scenarios {
		jsonData, err := YamlToJson(scenario.yaml)
		if scenario.expectedError == "" {
			assert.NoError(t, err, "Unexpected error in scenario %q", scenario.description)

			jsonBytes, err := json.Marshal(jsonData)
			assert.NoError(t, err, "Unexpected error in scenario %q", scenario.description)

			jsonString := string(jsonBytes)
			assert.Equal(t, scenario.expectedJson, jsonString, "Unexpected json in scenario %q", scenario.description)
		} else {
			assert.EqualError(t, err, scenario.expectedError, "Unexpected error in scenario %q", scenario.description)
		}
	}
}
