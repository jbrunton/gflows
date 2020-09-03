package yamlutil

import (
	"gopkg.in/yaml.v2"
)

func NormalizeWorkflow(content string) (string, error) {
	var yamlData yaml.MapSlice
	err := yaml.Unmarshal([]byte(content), &yamlData)
	if err != nil {
		return "", err
	}

	for i, item := range yamlData {
		if item.Key == true {
			// Unfortunately GitHub workflows use the "on" reserved word, which canonically is treated as true, as a top
			// level key. We therefore guess that any key that gets parsed as true is intended to be used for the "on" key.
			yamlData[i] = yaml.MapItem{Key: "on", Value: item.Value}
		}
	}

	result, err := yaml.Marshal(yamlData)
	return string(result), err
}
