package upgrade_tests

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/timescale/tobs/cli/pkg/helm"
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

func TestUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping upgrade tests")
	}

	installTobsRecentRelease()

	upgradeTobsLatest()

	fmt.Println("Successfully upgraded tobs to latest version")

	out := exec.Command("helm", "dep", "up", "./../testdata/chart1/")
	output, err := out.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		t.Fatal(err)
	}

	out = exec.Command("helm", "dep", "up", "./../testdata/chart2/")
	output, err = out.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		t.Fatal(err)
	}

	err = test_utils.DeleteWebhooks()
	if err != nil {
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart1/", "-f", "./../testdata/chart1/values.yaml", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	output, err = out.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart1/", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	k, err := out.CombinedOutput()
	if err != nil {
		if string(k) != "Error: failed to upgrade there is no latest helm chart available and existing helm deployment values are same as the provided values\n" {
			t.Fatal(err)
		}
	} else {
		err = errors.New("expected an error but didn't get an error")
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart1/", "-f", "./../testdata/chart1/values.yaml", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	k, err = out.CombinedOutput()
	if err != nil {
		if string(k) != "Error: failed to upgrade there is no latest helm chart available and existing helm deployment values are same as the provided values\n" {
			t.Fatal(err)
		}
	} else {
		err = errors.New("expected an error but didn't get an error")
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart2/", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y", "--same-chart")
	k, err = out.CombinedOutput()
	if err != nil {
		if string(k) != "Error: provided helm chart is newer compared to existing deployed helm chart cannot upgrade as --same-chart flag is provided\n" {
			t.Fatal(err)
		}
	} else {
		err = errors.New("expected an error but didn't get an error")
		t.Fatal(err)
	}

	err = test_utils.DeleteWebhooks()
	if err != nil {
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart2/", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	output, err = out.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		t.Fatal(err)
	}

	helmClient = helm.NewClient(NAMESPACE)
	defer helmClient.Close()
	chartDetails, err := helmClient.GetDeployedChartMetadata(RELEASE_NAME)
	if err != nil {
		t.Fatal(err)
	}
	if chartDetails.Chart != "tobs" && chartDetails.Version == "0.5.8" {
		t.Fatal("failed to verify expected chart version after upgrade", chartDetails.Chart)
	}

	err = test_utils.DeleteWebhooks()
	if err != nil {
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-f", "./../testdata/f4.yaml", "-c", "./../testdata/chart2/", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	output, err = out.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		t.Fatal(err)
	}
	size, err := test_utils.GetUpdatedPromscaleMemResource(RELEASE_NAME, NAMESPACE)
	if err != nil {
		t.Fatal(err)
	}
	if size != "1Gi" {
		t.Fatal("failed to validate expected promscale memory size from tobs upgrade")
	}
}
