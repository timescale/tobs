package helm_tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/timescale/tobs/cli/pkg/helm"
	"github.com/timescale/tobs/cli/pkg/utils"
)

var (
	helmClientTest      helm.Client
	oldNamespaceFromEnv string
	tobsVersion         string
	NAMESPACE           = "hello"
	CHART_NAME          = "timescale/tobs"
	LATEST_CHART        = "./../../../chart"
	PATH_TO_TOBS        = "./../../bin/tobs"
	PATH_TO_TEST_VALUES = "./../testdata/main-values.yaml"
	PATH_TO_MAIN_VALUES = "./../../../chart/values.yaml"
	DEFAULT_TOBS_NAME   = "tobs"
)

func testNewHelmClient() {
	opt := &helm.ClientOptions{
		Namespace: NAMESPACE,
		Linting:   true,
	}

	testSetNamespaceInEnv(NAMESPACE)

	var err error
	helmClientTest, err = helm.New(opt)
	if err != nil {
		log.Fatal(err)
	}
}

// setEnvSettings sets the client's environment settings
func testSetNamespaceInEnv(namespace string) {
	err := os.Setenv("HELM_NAMESPACE", "hii")
	if err != nil {
		log.Fatal("failed to set HELM_NAMESPACE env variable", err)
	}

	// capture the old namespace from env vars
	oldNamespaceFromEnv = os.Getenv("HELM_NAMESPACE")

	// set the namespace with this ugly workaround because cli.EnvSettings.namespace is private
	// thank you helm!
	err = os.Setenv("HELM_NAMESPACE", namespace)
	if err != nil {
		log.Fatalf("failed to set HELM_NAMESPACE env variable %v", err)
	}
}

// to make sure we don't leave unwanted envs set.
func testUnSetHelmNamespaceEnv() {
	err := os.Setenv("HELM_NAMESPACE", oldNamespaceFromEnv)
	if err != nil {
		fmt.Printf("Error: failed to unset HELM_NAMESPACE env variable to '' : %v\n", err)
	}
}

// add or update tobs helm repo
func testHelmClientAddOrUpdateChartRepoPublic() {
	// Add a chart-repository to the client
	if err := helmClientTest.AddOrUpdateChartRepo(utils.DEFAULT_REGISTRY_NAME, utils.REPO_LOCATION); err != nil {
		log.Fatal(err)
	}
}

// install or upgrade tobs instllation
func testHelmClientInstallOrUpgradeChart() {
	log.Println("Installing Tobs secrets...")
	runTsdb := exec.Command(PATH_TO_TOBS, "install", "--namespace", NAMESPACE, "--only-secrets")
	_, err := runTsdb.CombinedOutput()
	if err != nil {
		log.Fatalf("Error installing tobs secrets %v:", err)
	}

	// Define the chart to be installed
	chartSpec := helm.ChartSpec{
		ReleaseName:     DEFAULT_TOBS_NAME,
		ChartName:       CHART_NAME,
		Namespace:       NAMESPACE,
		Version:         tobsVersion,
		CreateNamespace: true,
		ValuesFiles:     []string{PATH_TO_TEST_VALUES},
	}

	log.Println("Installing Tobs...")
	res, err := helmClientTest.InstallOrUpgradeChart(context.Background(), &chartSpec)
	if err != nil {
		log.Fatal(err)
	}

	if res.Info.Status != "deployed" {
		log.Fatal("failed to perform helm chart install")
	}
}

// list tobs deployment and verify it post uninstall
func testTobsReleasePostUninstall() {
	res, err := helmClientTest.GetDeployedChartMetadata("tobs", NAMESPACE)
	if err == nil && res.Name == DEFAULT_TOBS_NAME {
		log.Fatal("the tobs release after uninstalling are still showing up....", res)
	}
	testUnSetHelmNamespaceEnv()

	ns := os.Getenv("HELM_NAMESPACE")
	if ns != oldNamespaceFromEnv || ns != "hii" {
		log.Fatal("failed to set back old HELM_NAMESPACE value to env variable")
	}
}

func testGetChartMetadata() {
	chart, err := helmClientTest.GetChartMetadata(CHART_NAME)
	if err != nil {
		log.Fatal(err)
	}

	// assign the latest chart version from timescale/tobs helm chart
	tobsVersion = chart.Version
	if chart.Name != DEFAULT_TOBS_NAME {
		log.Fatalf("failed to verify chart metadata %v", chart)
	}
}

func TestExportValueFromChart(t *testing.T) {
	res, err := helmClientTest.ExportValuesFieldFromChart(CHART_NAME, PATH_TO_TEST_VALUES, []string{"promscale", "resources", "requests", "memory"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := res.(string)
	if !ok {
		t.Fatal("failed to get expected value string from export chart value field")
	}
	if v != "50Mi" {
		t.Fatal("failed to verify exportChartValue")
	}
}

func TestGetChartValues(t *testing.T) {
	res, err := helmClientTest.GetChartValues(LATEST_CHART)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := ioutil.ReadFile(PATH_TO_MAIN_VALUES)
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.Compare(res, expected)

	if r != 0 {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(string(res), string(expected), false)
		t.Fatalf("failed to verify get chart values:\n%v", dmp.DiffPrettyText(diffs))
	}
}

func TestGetDeployedChartMetadata(t *testing.T) {
	chart, err := helmClientTest.GetDeployedChartMetadata(DEFAULT_TOBS_NAME, NAMESPACE)
	if err != nil {
		t.Fatal(err)
	}

	if chart.Name != DEFAULT_TOBS_NAME && chart.Namespace != NAMESPACE && chart.Status != "deployed" && chart.Version != tobsVersion {
		t.Fatalf("failed to verify deployed chart metadata %v", chart)
	}
}

func TestHelmClientGetReleaseValues(t *testing.T) {
	_, err := helmClientTest.GetReleaseValues(DEFAULT_TOBS_NAME)
	if err != nil {
		t.Fatal("failed to get all release values", err)
	}
}

func TestExportValueFromRelease(t *testing.T) {
	res, err := helmClientTest.ExportValuesFieldFromRelease(DEFAULT_TOBS_NAME, []string{"promscale", "resources", "requests", "cpu"})
	if err != nil {
		t.Fatal(err)
	}

	v, ok := res.(string)
	if !ok {
		t.Fatal("failed to get expected value string from export chart value field")
	}
	if v != "10m" {
		t.Fatal("failed to verify exportChartValueFromRelease")
	}
}

func TestHelmClientUninstallRelease(t *testing.T) {
	// Define the released chart to be installed
	chartSpec := helm.ChartSpec{
		ReleaseName: DEFAULT_TOBS_NAME,
		ChartName:   CHART_NAME,
		Namespace:   NAMESPACE,
	}

	testNewHelmClient()
	if err := helmClientTest.UninstallRelease(&chartSpec); err != nil {
		t.Fatal(err)
	}
}
