package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplateArrayProperty(t *testing.T) {
	config, _ := parseConfig([]byte(strings.Join([]string{
		"templates:",
		"  engine: ytt",
		"  defaults:",
		"    jsonnet:",
		"      jpath:",
		"        - vendor",
		"  overrides:",
		"    my-workflow:",
		"      jsonnet:",
		"        jpath:",
		"        - my-lib",
	}, "\n")))
	assert.Equal(t, []string{"vendor"}, config.GetTemplateArrayProperty("some-workflow", func(config *GFlowsTemplateConfig) []string {
		return config.Jsonnet.JPath
	}))
	assert.Equal(t, []string{"my-lib", "vendor"}, config.GetTemplateArrayProperty("my-workflow", func(config *GFlowsTemplateConfig) []string {
		return config.Jsonnet.JPath
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
