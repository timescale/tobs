package upgrade_tests

import (
	"errors"
	"fmt"
	"github.com/timescale/tobs/cli/pkg/helm"
	"os/exec"
	"testing"

	"github.com/timescale/tobs/cli/pkg/k8s"
)

func TestUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping upgrade tests")
	}

	installTobsRecentRelease()

	upgradeTobsLatest()

	fmt.Println("Successfully upgraded tobs to latest version")

	fmt.Println("deleting webhooks 1")
	kubeClient.K8s, _ = k8s.NewClient()
	err := kubeClient.DeleteWebhooks()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("deleted webhooks, performing upgrade 2")

	out := exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart1/", "-f", "./../testdata/chart1/values.yaml", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	output, err := out.CombinedOutput()
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

	fmt.Println("deleting webhooks 2")
	kubeClient.K8s, _ = k8s.NewClient()
	err = kubeClient.DeleteWebhooks()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("performing upgrade 3")
	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart2/", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	output, err = out.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		t.Fatal(err)
	}

	hc = helm.NewClient(NAMESPACE)
	chartDetails, err := hc.GetDeployedChartMetadata(RELEASE_NAME)
	if err != nil {
		t.Fatal(err)
	}
	if chartDetails.Chart != "tobs" && chartDetails.Version == "0.5.8" {
		t.Fatal("failed to verify expected chart version after upgrade", chartDetails.Chart)
	}

	fmt.Println("deleting webhooks 3")
	kubeClient.K8s, _ = k8s.NewClient()
	err = kubeClient.DeleteWebhooks()
	if err != nil {
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-f", "./../testdata/f4.yaml", "-c", "./../testdata/chart2/", "--namespace", NAMESPACE, "--name", RELEASE_NAME, "-y")
	output, err = out.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		t.Fatal(err)
	}

	fmt.Println("Checking resources after upgrade")
	t.Log("HIII HELLO")
	kubeClient.K8s, _ = k8s.NewClient()
	size, err := kubeClient.GetUpdatedPromscaleMemResource(RELEASE_NAME, NAMESPACE)
	if err != nil {
		t.Fatal(err)
	}
	if size != "1Gi" {
		t.Fatal("failed to validate expected promscale memory size from tobs upgrade")
	}
}
