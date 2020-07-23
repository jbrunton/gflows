package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanComments(t *testing.T) {
	assert.Equal(t, "foo: bar", CleanComments("foo: bar"))
	assert.Equal(t, "\nfoo: bar", CleanComments("# some comment\nfoo: bar"))
}
