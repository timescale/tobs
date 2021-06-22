package helm

import (
	"reflect"
	"testing"
)

func TestExportValuesFieldFromChart(t *testing.T) {
	type args struct {
		chart string
		keys  []string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Success test-case to fetch nothing as the provided keys are empty",
			args: args{
				chart: "../../../chart/",
				keys:  nil,
			},
			want: nil,
		},
		{
			name: "Success test-case to fetch timescaledb-single.loadBalancer,enabled",
			args: args{
				chart: "../../../chart/",
				keys:  []string{"timescaledb-single", "loadBalancer", "enabled"},
			},
			want: false,
		},
		{
			name: "Success test-case to fetch timescaledb-single.backup.enabled",
			args: args{
				chart: "../../../chart/",
				keys:  []string{"timescaledb-single", "backup", "enabled"},
			},
			want: false,
		},
		{
			name: "Failure test-case as provided keys are invalid",
			args: args{
				chart: "../../../chart/",
				keys:  []string{"timescaledb-single", "backup", "enab**"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	helmClient := NewClient("default")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := helmClient.ExportValuesFieldFromChart(tt.args.chart, "", tt.args.keys)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportValuesFieldValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExportValuesFieldValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}