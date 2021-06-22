package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	REPO_LOCATION         = "https://charts.timescale.com"
	DEFAULT_CHART         = "timescale/tobs"
	DEFAULT_REGISTRY_NAME = "timescale"
	UpgradeJob_040        = "tobs-prometheus-permission-change"
	PrometheusPVCName     = "prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"
	Version_040           = "0.4.0"
)

func ParseVersion(s string, width int) (int64, error) {
	strList := strings.Split(s, ".")
	format := fmt.Sprintf("%%s%%0%ds", width)
	v := ""
	for _, value := range strList {
		v = fmt.Sprintf(format, v, value)
	}
	var result int64
	var err error
	if result, err = strconv.ParseInt(v, 10, 64); err != nil {
		return 0, fmt.Errorf("failed: parseVersion(%s): error=%s", s, err)
	}
	return result, nil
}

func ConfirmAction() {
	fmt.Print("confirm the action by typing y or yes and press enter: ")
	for {
		confirm := ""
		_, err := fmt.Scan(&confirm)
		if err != nil {
			log.Fatalf("couldn't get confirmation from the user %v", err)
		}
		confirm = strings.TrimSuffix(confirm, "\n")
		if confirm == "yes" || confirm == "y" {
			return
		} else {
			fmt.Println("confirmation doesn't match with expected key. please type \"y\" or \"yes\" and press enter\nHint: Press (ctrl+c) to exit")
		}
	}
}

func ErrorTobsDeploymentNotFound(releaseName string) error {
	return fmt.Errorf("unable to find the tobs deployment with name %s", releaseName)
}

func GetTimescaleDBLabels(releaseName string) map[string]string {
	return map[string]string{
		"app":     releaseName + "-timescaledb",
		"release": releaseName,
	}
}

func GetPrometheusLabels() map[string]string {
	return map[string]string{
		"app":        "prometheus",
		"prometheus": "tobs-kube-prometheus-prometheus",
	}
}

func GetTimescaleDBsecretLabels(releaseName string) map[string]string {
	return map[string]string{
		"app":          releaseName + "-timescaledb",
		"cluster-name": releaseName,
	}
}