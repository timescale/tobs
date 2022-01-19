package helm

import (
	"encoding/json"
	"reflect"
	"testing"
)

var (
	valuesYaml = "./../../tests/testdata/helm-unit-values.yaml"
)

func TestChartSpec_GetValuesMap(t *testing.T) {
	type fields struct {
		ValuesYaml  string
		ValuesFiles []string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Add a top-level section",
			fields: fields{
				ValuesYaml: `
timescaledb-single:
  enabled: false`,
				ValuesFiles: []string{valuesYaml},
			},
			want: `{
				"promscale": {
					"enabled": true,
					"resources": {
						"requests": {
							"memory": "50Mi",
							"cpu":    "10m"
						}
					}
				},
				"timescaledb-single": {
					"enabled": false
				}
			}`,
			wantErr: false,
		},
		{
			name: "Modify nested section",
			fields: fields{
				ValuesYaml: `
promscale:
  resources:
    requests:
      cpu: 50m
      memory: 500Mi`,
				ValuesFiles: []string{valuesYaml},
			},
			want: `{
				"promscale": {
					"enabled": true,
					"resources": {
						"requests": {
							"memory": "500Mi",
							"cpu":    "50m"
						}
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "Merging an invalid field",
			fields: fields{
				ValuesYaml:  "",
				ValuesFiles: []string{valuesYaml},
			},
			wantErr: false,
			want: `{
				"promscale": {
					"enabled": true,
					"resources": {
						"requests": {
							"memory": "50Mi",
							"cpu":    "10m"
						}
					}
				}
			}`,
		},
		{
			name: "Providing a invalid values file path",
			fields: fields{
				ValuesYaml:  "",
				ValuesFiles: []string{"abc"},
			},
			wantErr: true,
			want:    "null",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ChartSpec{
				ValuesYaml:  tt.fields.ValuesYaml,
				ValuesFiles: tt.fields.ValuesFiles,
			}
			got, err := spec.GetValuesMap()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValuesMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var want map[string]interface{}
			err = json.Unmarshal([]byte(tt.want), &want)
			if err != nil {
				t.Errorf("Dataset unmarshaling failed: %v", err)
				return
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("GetValuesMap() \ngot  = %v\nwant = %v", got, want)
			}
		})
	}
}
