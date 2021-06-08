package helm

import (
	"fmt"
)

type ChartMetadata struct {
	APIVersion   string `yaml:"apiVersion"`
	AppVersion   string `yaml:"appVersion"`
	Dependencies []struct {
		Condition  string `yaml:"condition"`
		Name       string `yaml:"name"`
		Repository string `yaml:"repository"`
		Version    string `yaml:"version"`
	} `yaml:"dependencies"`
	Description string `yaml:"description"`
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Version     string `yaml:"version"`
}

type DeployedChartMetadata struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
	Version    string
}

// Fetches the value from provided keys
func FetchValue(f interface{}, keys []string) (interface{}, error) {
	if keys == nil {
		return nil, nil
	}

	itemsMap := f.(map[string]interface{})
	for k, v := range itemsMap {
		if keys != nil && k == keys[0] {
			if len(keys[1:]) == 0 {
				return v, nil
			}
			v1, _ := FetchValue(v, keys[1:])
			if v1 != nil {
				return v1, nil
			}
		}
	}
	return nil, fmt.Errorf("failed to find the value from the keys in values.yaml %v", keys)
}

// ConvertMapI2MapS walks the given dynamic object recursively, and
// converts maps with interface{} key type to maps with string key type.
// This function comes handy if you want to marshal a dynamic object into
// JSON where maps with interface{} key type are not allowed.
//
// Recursion is implemented into values of the following types:
//   -map[interface{}]interface{}
//   -map[string]interface{}
//   -[]interface{}
//
// When converting map[interface{}]interface{} to map[string]interface{},
// fmt.Sprint() with default formatting is used to convert the key to a string key.
func ConvertMapI2MapS(v interface{}) interface{} {
	switch x := v.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v2 := range x {
			switch k2 := k.(type) {
			case string: // Fast check if it's already a string
				m[k2] = ConvertMapI2MapS(v2)
			default:
				m[fmt.Sprint(k)] = ConvertMapI2MapS(v2)
			}
		}
		v = m

	case []interface{}:
		for i, v2 := range x {
			x[i] = ConvertMapI2MapS(v2)
		}

	case map[string]interface{}:
		for k, v2 := range x {
			x[k] = ConvertMapI2MapS(v2)
		}
	}

	return v
}
