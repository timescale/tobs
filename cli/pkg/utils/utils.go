package utils

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/timescale/tobs/cli/pkg/k8s"
)

const (
	REPO_LOCATION         = "https://charts.timescale.com"
	DEFAULT_CHART         = "timescale/tobs"
	DEFAULT_REGISTRY_NAME = "timescale"
	UpgradeJob_040        = "tobs-prometheus-permission-change"
	PrometheusPVCName     = "prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"
	Version_040           = "0.4.0"
)

func ErrorTobsDeploymentNotFound(releaseName string) error {
	return fmt.Errorf("unable to find the tobs deployment with name %s", releaseName)
}

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

func GetTimescaleDBURI(k8sClient k8s.Client, namespace, name string) (string, error) {
	secretName := name + "-timescaledb-uri"
	secrets, err := k8sClient.KubeGetAllSecrets(namespace)
	if err != nil {
		return "", err
	}

	for _, s := range secrets.Items {
		if s.Name == secretName {
			if bytepass, exists := s.Data["db-uri"]; exists {
				uriData := string(bytepass)
				return uriData, nil
			} else {
				// found the secret but failed to find the value with indexed key.
				return "", fmt.Errorf("could not get TimescaleDB URI with secret key index as db-uri from %s", secretName)
			}
		}
	}

	return "", nil
}

func GetDBPassword(k8sClient k8s.Client, secretKey, name, namespace string) ([]byte, error) {
	secret, err := k8sClient.KubeGetSecret(namespace, name+"-credentials")
	if err != nil {
		return nil, fmt.Errorf("could not get TimescaleDB password: %w", err)
	}

	if bytepass, exists := secret.Data[secretKey]; exists {
		return bytepass, nil
	}

	return nil, fmt.Errorf("user not found")
}

func GetTimescaleDBsecretLabels(releaseName string) map[string]string {
	return map[string]string{
		"app":          releaseName + "-timescaledb",
		"cluster-name": releaseName,
	}
}
