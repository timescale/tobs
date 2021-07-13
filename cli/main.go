/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"github.com/timescale/tobs/cli/cmd"
	_ "github.com/timescale/tobs/cli/cmd/grafana"
	_ "github.com/timescale/tobs/cli/cmd/helm"
	_ "github.com/timescale/tobs/cli/cmd/install"
	_ "github.com/timescale/tobs/cli/cmd/metrics"
	_ "github.com/timescale/tobs/cli/cmd/port-forward"
	_ "github.com/timescale/tobs/cli/cmd/prometheus"
	_ "github.com/timescale/tobs/cli/cmd/promlens"
	_ "github.com/timescale/tobs/cli/cmd/promscale"
	_ "github.com/timescale/tobs/cli/cmd/timescaledb"
	_ "github.com/timescale/tobs/cli/cmd/uninstall"
	_ "github.com/timescale/tobs/cli/cmd/upgrade"
	_ "github.com/timescale/tobs/cli/cmd/version"
	_ "github.com/timescale/tobs/cli/cmd/volume"
)

func main() {
	cmd.Execute()
}
