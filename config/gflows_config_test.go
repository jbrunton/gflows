package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplateArrayProperty(t *testing.T) {
	config, _ := parseConfig([]byte(strings.Join([]string{
		"templates:",
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
