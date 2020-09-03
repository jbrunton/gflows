package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeWorkflow(t *testing.T) {
	result, _ := NormalizeWorkflow("foo: bar")
	assert.Equal(t, "foo: bar\n", result)

	result, _ = NormalizeWorkflow("on: foo")
	assert.Equal(t, "\"on\": foo\n", result)
}
