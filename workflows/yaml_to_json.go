package workflows

import "fmt"

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
			if keyPrefix == "" && key == true {
				// Unfortunately GitHub workflows use the "on" reserved word, which canonically is treated as true, as a top
				// level key. We therefore guess that any key that gets parsed as true is intended to be used for the "on" key.
				str = "on"
			} else {
				return nil, formatInvalidKeyError(keyPrefix, key)
			}
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
