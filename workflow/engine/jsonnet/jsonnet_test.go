package jsonnet

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	testCases := []struct {
		jsonInput       string
		expectedJsonnet string
	}{
		{
			jsonInput:       "{}",
			expectedJsonnet: "{}",
		},
		{
			jsonInput: `{
				"foo": "bar",
				"baz": 123
			}`,
			expectedJsonnet: strings.Join([]string{
				`{`,
				`  baz: 123,`,
				`  foo: "bar"`,
				`}`,
			}, "\n"),
		},
		{
			jsonInput: `{
				"foo": "bar",
				"if": "keyword",
				"key-with-dashes": 123,
				"key with spaces": 456
			}`,
			expectedJsonnet: strings.Join([]string{
				`{`,
				`  foo: "bar",`,
				`  "if": "keyword",`,
				`  "key with spaces": 456,`,
				`  "key-with-dashes": 123`,
				`}`,
			}, "\n"),
		},
		{
			jsonInput: `{
				"nested": {
					"foo": "bar"
				},
				"array": [1, 2]
			}`,
			expectedJsonnet: strings.Join([]string{
				`{`,
				`  array: [`,
				`    1,`,
				`    2`,
				`  ],`,
				`  nested: {`,
				`    foo: "bar"`,
				`  }`,
				`}`,
			}, "\n"),
		},
	}

	for _, testCase := range testCases {
		v := make(map[string]interface{})
		json.Unmarshal([]byte(testCase.jsonInput), &v)
		out, err := Marshal(v)
		assert.Equal(t, testCase.expectedJsonnet, string(out))
		assert.NoError(t, err)
	}
}
