package jsonnet

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJsonnetEmpty(t *testing.T) {
	v := make(map[string]interface{})
	json.Unmarshal([]byte{}, &v)
	out, err := MarshalJsonnet(v)
	assert.Equal(t, "{}", string(out))
	assert.NoError(t, err)
}

func TestMarshalJsonnetSimple(t *testing.T) {
	v := make(map[string]interface{})
	json.Unmarshal([]byte(`{"foo": "bar", "baz": 123}`), &v)
	out, err := MarshalJsonnet(v)
	expected := strings.Join([]string{
		`{`,
		`  baz: 123,`,
		`  foo: "bar"`,
		`}`,
	}, "\n")
	assert.Equal(t, expected, string(out))
	assert.NoError(t, err)
}

func TestMarshalJsonnetKeywords(t *testing.T) {
	v := make(map[string]interface{})
	json.Unmarshal([]byte(`{"foo": "bar", "if": "${{ github.expression }}" }`), &v)
	out, err := MarshalJsonnet(v)
	expected := strings.Join([]string{
		`{`,
		`  foo: "bar",`,
		`  "if": "${{ github.expression }}"`,
		`}`,
	}, "\n")
	assert.Equal(t, expected, string(out))
	assert.NoError(t, err)
}

func TestMarshalJsonnetComplexKeys(t *testing.T) {
	v := make(map[string]interface{})
	json.Unmarshal([]byte(`{"key-with-dashes": 123,"key with spaces": 456}`), &v)
	out, err := MarshalJsonnet(v)
	expected := strings.Join([]string{
		`{`,
		`  "key with spaces": 456,`,
		`  "key-with-dashes": 123`,
		`}`,
	}, "\n")
	assert.Equal(t, expected, string(out))
	assert.NoError(t, err)
}
