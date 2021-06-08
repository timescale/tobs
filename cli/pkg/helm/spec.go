package helm

import (
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"sigs.k8s.io/yaml"
)

// GetValuesMap returns the mapped out values of a chart
func (spec *ChartSpec) GetValuesMap() (map[string]interface{}, error) {
	var values map[string]interface{}

	// unmarshal the string to YAML
	err := yaml.Unmarshal([]byte(spec.ValuesYaml), &values)
	if err != nil {
		return nil, err
	}

	valueOpts := ValuesOptions{
		ValueFiles:         spec.ValuesFiles,
		ValuesYamlIndented: values,
	}

	p := getter.All(cli.New())
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	return vals, nil
}