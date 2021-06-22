package helm

import (
	"fmt"
	"github.com/timescale/tobs/cli/pkg/utils"
	"io/ioutil"
	"sigs.k8s.io/yaml"
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

func (c *ClientInfo) ExportValuesFieldFromRelease(releaseName string, keys []string) (interface{}, error) {
	res, err := c.GetAllReleaseValues(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to do helm get values from the helm release %w", err)
	}
	return FetchValue(res, keys)
}

func (c *ClientInfo) GetDeployedChartMetadata(releaseName string) (*DeployedChartMetadata, error) {
	var charts []DeployedChartMetadata
	l, err := c.ListDeployedReleases()
	if err != nil {
		return nil, err
	}
	for _, k := range l {
		d := DeployedChartMetadata{
			Name:       k.Name,
			Namespace:  k.Namespace,
			Updated:    k.Info.LastDeployed.String(),
			Status:     k.Info.Status.String(),
			Chart:      k.Chart.Name(),
			AppVersion: k.Chart.Metadata.AppVersion,
			Version:    k.Chart.Metadata.Version,
		}
		charts = append(charts, d)
	}

	for _, c := range charts {
		if c.Name == releaseName {
			return &c, nil
		}
	}

	return nil, utils.ErrorTobsDeploymentNotFound(releaseName)
}

func (c *ClientInfo) GetTobsChartMetadata(chart string) (*ChartMetadata, error) {
	chartDetails := &ChartMetadata{}
	res, err := c.InspectChartYaml(chart)
	if err != nil {
		return chartDetails, fmt.Errorf("failed to inspect chart %w", err)
	}

	err = yaml.Unmarshal(res, chartDetails)
	if err != nil {
		return chartDetails, fmt.Errorf("failed to unmarshal helm chart metadata %w", err)
	}

	return chartDetails, nil
}

func (c *ClientInfo) ExportValuesFieldFromChart(chart string, customValuesFile string, keys []string) (interface{}, error) {
	res, err := c.GetValuesYamlFromChart(chart, customValuesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to do helm show values on the helm chart %w", err)
	}
	return FetchValue(res, keys)
}

func (c *ClientInfo) GetValuesYamlFromChart(chart, file string) (interface{}, error) {
	var res []byte
	var err error
	if file != "" {
		res, err = ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to read values from provided file %w", err)
		}
	} else {
		res, err = c.GetChartValues(chart)
		if err != nil {
			return nil, err
		}
	}

	var i interface{}
	err = yaml.Unmarshal(res, &i)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal existing values.yaml file %w", err)
	}

	values := ConvertMapI2MapS(i)
	return values, nil
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
