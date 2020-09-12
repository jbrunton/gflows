package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jbrunton/gflows/io"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplateArrayProperty(t *testing.T) {
	config, _ := parseConfig([]byte(strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"  defaults:",
		"    libs:",
		"      - vendor",
		"  overrides:",
		"    my-workflow:",
		"      libs:",
		"      - my-lib",
	}, "\n")))
	assert.Equal(t, []string{"vendor"}, config.GetTemplateArrayProperty("some-workflow", func(config *GFlowsTemplateConfig) []string {
		return config.Libs
	}))
	assert.Equal(t, []string{"vendor", "my-lib"}, config.GetTemplateArrayProperty("my-workflow", func(config *GFlowsTemplateConfig) []string {
		return config.Libs
	}))
}

func TestGetWorkflowBoolProperty(t *testing.T) {
	config, _ := parseConfig([]byte(strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"workflows:",
		"  defaults:",
		"    checks:",
		"      schema:",
		"        enabled: true",
		"  overrides:",
		"    my-workflow:",
		"      checks:",
		"        schema:",
		"          enabled: false",
		"        content:",
		"          enabled: true",
	}, "\n")))

	schemaEnabledSelector := func(config *GFlowsWorkflowConfig) *bool {
		return config.Checks.Schema.Enabled
	}
	contentEnabledSelector := func(config *GFlowsWorkflowConfig) *bool {
		return config.Checks.Content.Enabled
	}

	scenarios := []struct {
		workflowName   string
		defaultValue   bool
		expectedResult bool
		selector       func(*GFlowsWorkflowConfig) *bool
	}{
		{
			workflowName:   "some-workflow",
			defaultValue:   true,
			expectedResult: true,
			selector:       schemaEnabledSelector,
		},
		{
			workflowName:   "some-workflow",
			defaultValue:   false,
			expectedResult: true,
			selector:       schemaEnabledSelector,
		},
		{
			workflowName:   "my-workflow",
			defaultValue:   true,
			expectedResult: false,
			selector:       schemaEnabledSelector,
		},
		{
			workflowName:   "my-workflow",
			defaultValue:   true,
			expectedResult: false,
			selector:       schemaEnabledSelector,
		},
		{
			workflowName:   "some-workflow",
			defaultValue:   true,
			expectedResult: true,
			selector:       contentEnabledSelector,
		},
		{
			workflowName:   "some-workflow",
			defaultValue:   false,
			expectedResult: false,
			selector:       contentEnabledSelector,
		},
		{
			workflowName:   "my-workflow",
			defaultValue:   true,
			expectedResult: true,
			selector:       contentEnabledSelector,
		},
		{
			workflowName:   "my-workflow",
			defaultValue:   false,
			expectedResult: true,
			selector:       contentEnabledSelector,
		},
	}

	for _, scenario := range scenarios {
		actualResult := config.GetWorkflowBoolProperty(scenario.workflowName, scenario.defaultValue, scenario.selector)
		assert.Equal(t, scenario.expectedResult, actualResult, "Failures for scenario %+v", scenario)
	}
}

func TestValidateConfig(t *testing.T) {
	scenarios := []struct {
		description    string
		config         string
		expectedError  string
		expectedOutput string
	}{
		{
			description: "invalid templates",
			config: strings.Join([]string{
				"templates:",
				"  foo: bar",
			}, "\n"),
			expectedError:  "invalid config",
			expectedOutput: "Schema error: templates: Additional property foo is not allowed\n",
		},
		{
			description: "invalid templates override",
			config: strings.Join([]string{
				"templates:",
				"  overrides:",
				"    my-workflow:",
				"      libs: [my-lib]",
				"      foo: bar",
			}, "\n"),
			expectedError:  "invalid config",
			expectedOutput: "Schema error: templates.overrides.my-workflow: Additional property foo is not allowed\n",
		},
		{
			description: "invalid workflows override",
			config: strings.Join([]string{
				"workflows:",
				"  overrides:",
				"    my-workflow:",
				"      foo: bar",
			}, "\n"),
			expectedError:  "invalid config",
			expectedOutput: "Schema error: workflows.overrides.my-workflow: Additional property foo is not allowed\n",
		},
		{
			description: "valid config",
			config: strings.Join([]string{
				"templates:",
				"  engine: ytt",
				"  defaults:",
				"    libs: [vendor]",
				"  overrides:",
				"    my-workflow:",
				"      libs: [my-lib]",
				"workflows:",
				"  defaults:",
				"    checks:",
				"      schema:",
				"        enabled: true",
				"        uri: example.com",
				"      content:",
				"        enabled: true",
				"  overrides:",
				"    my-workflow:",
				"      checks:",
				"        content:",
				"          enabled: false",
			}, "\n"),
			expectedError:  "",
			expectedOutput: "",
		},
	}

	for _, scenario := range scenarios {
		out := new(bytes.Buffer)
		logger := io.NewLogger(out, false, false)
		err := validateConfig(scenario.config, logger)

		if scenario.expectedError == "" {
			assert.NoError(t, err, "Unexpected error in scenario %s", scenario.description)
		} else {
			assert.EqualError(t, err, scenario.expectedError, "Unexpected error in scenario %q", scenario.description)
		}
		assert.Equal(t, scenario.expectedOutput, out.String(), "Output mismatch in scenario %q", scenario.description)
	}
}
