package tests

import (
	"github.com/timescale/tobs/cli/pkg/k8s"
	"os/exec"
	"strings"
	"testing"
)

func testVolumeExpansion(t testing.TB, timescaleDBStorage, timescaleDBWal, prometheusStorage string) {
	cmds := []string{"volume", "expand", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if timescaleDBStorage != "" {
		cmds = append(cmds, "--timescaleDB-storage", timescaleDBStorage)
	}

	if timescaleDBWal != "" {
		cmds = append(cmds, "--timescaleDB-wal", timescaleDBWal)
	}

	if prometheusStorage != "" {
		cmds = append(cmds, "--prometheus-storage", prometheusStorage)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	 expand := exec.Command("./../bin/tobs", cmds...)
	 out, err := expand.CombinedOutput()
	 if err != nil {
		 t.Logf(string(out))
		 t.Fatal(err)
	 }
}

func testVolumeGet(t testing.TB, timescaleDBStorage, timescaleDBWal, prometheusStorage bool) {
	cmds := []string{"volume", "get", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if timescaleDBStorage {
		cmds = append(cmds, "--timescaleDB-storage")
	}

	if timescaleDBWal {
		cmds = append(cmds, "--timescaleDB-wal")
	}

	if prometheusStorage {
		cmds = append(cmds, "--prometheus-storage")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	expand := exec.Command("./../bin/tobs", cmds...)
	out, err := expand.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}


func TestVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Prometheus tests")
	}

	testVolumeGet(t, true, true, true)
	testVolumeGet(t, false, true, true)
	testVolumeGet(t, true, true, false)
	testVolumeGet(t, true, false, false)
	testVolumeGet(t, false, false, true)

	// update default storageClass in Kind to allow pvc expansion
	err := k8s.UpdateStorageClassAllowVolumeExpand()
	if err != nil {
		t.Fatal(err)
	}

	testVolumeExpansion(t, "151Gi", "21Gi", "9Gi")
	testVolumeExpansion(t, "152Gi", "22Gi", "")
	testVolumeExpansion(t, "153Gi", "", "")
	testVolumeExpansion(t, "", "23Gi", "")
	testVolumeExpansion(t, "", "24Gi", "10Gi")
	testVolumeExpansion(t, "", "", "11Gi")
}
