package tobs_cli_tests

import (
	"errors"
	"fmt"
	"github.com/timescale/tobs/cli/tests/test-utils"
	"os/exec"
	"testing"

	"github.com/timescale/tobs/cli/pkg/utils"
)

func TestUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping upgrade tests")
	}

	out := exec.Command("helm", "dep", "up", "./../testdata/chart1/")
	_, err := out.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	out = exec.Command("helm", "dep", "up", "./../testdata/chart2/")
	_, err = out.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart1/", "-f", "./../testdata/chart1/values.yaml", "--namespace", "ns", "--name", "gg", "-y")
	_, err = out.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart1/", "--namespace", "ns", "--name", "gg", "-y")
	k, err := out.CombinedOutput()
	if err != nil {
		if string(k) != "Error: failed to upgrade there is no latest helm chart available and existing helm deployment values are same as the provided values\n" {
			t.Fatal(err)
		}
	} else {
		err = errors.New("expected an error but didn't get an error")
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart1/", "-f", "./../testdata/chart1/values.yaml", "--namespace", "ns", "--name", "gg", "-y")
	k, err = out.CombinedOutput()
	if err != nil {
		if string(k) != "Error: failed to upgrade there is no latest helm chart available and existing helm deployment values are same as the provided values\n" {
			t.Fatal(err)
		}
	} else {
		err = errors.New("expected an error but didn't get an error")
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart2/", "--namespace", "ns", "--name", "gg", "-y", "--same-chart")
	k, err = out.CombinedOutput()
	if err != nil {
		if string(k) != "Error: provided helm chart is newer compared to existing deployed helm chart cannot upgrade as --same-chart flag is provided\n" {
			t.Fatal(err)
		}
	} else {
		err = errors.New("expected an error but didn't get an error")
		t.Fatal(err)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-c", "./../testdata/chart2/", "--namespace", "ns", "--name", "gg", "-y")
	_, err = out.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	chartDetails, err := utils.GetDeployedChartMetadata(RELEASE_NAME, NAMESPACE)
	if err != nil {
		t.Fatal(err)
	}
	if chartDetails.Chart != "tobs-0.5.8" {
		t.Fatal("failed to verify expected chart version after upgrade", chartDetails.Chart)
	}

	out = exec.Command(PATH_TO_TOBS, "upgrade", "-f", "./../testdata/f6.yaml", "-c", "./../testdata/chart2/", "--namespace", "ns", "--name", "gg", "-y")
	k, err = out.CombinedOutput()
	if err != nil {
		fmt.Println(string(k))
		t.Fatal(err)
	}
	size, err := test_utils.GetUpdatedPromscaleMemResource()
	if err != nil {
		t.Fatal(err)
	}
	if size != "4Gi" {
		t.Fatal("failed to validate expected promscale memory size from tobs upgrade")
	}
}
