package tests

import (
	"errors"
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

func testVolumeGet(t testing.TB, timescaleDBStorage, timescaleDBWal, prometheusStorage bool) string {
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

	return string(out)
}


func TestVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Prometheus tests")
	}

	test1 := "PVC's of storage-volume\nExisting size of PVC: storage-volume-gg-timescaledb-0 is 150Gi\n\nPVC's of wal-volume\nExisting size of PVC: wal-volume-gg-timescaledb-0 is 20Gi\n\nPVC's of gg-prometheus-server\nExisting size of PVC: gg-prometheus-server is 8Gi\n\n"
	outputString := testVolumeGet(t, true, true, true)
	if outputString != test1 {
		t.Fatal(errors.New("failed to verify volume get test-1"))
	}

	test2 := "PVC's of wal-volume\nExisting size of PVC: wal-volume-gg-timescaledb-0 is 20Gi\n\nPVC's of gg-prometheus-server\nExisting size of PVC: gg-prometheus-server is 8Gi\n\n"
	outputString = testVolumeGet(t, false, true, true)
	if outputString != test2 {
		t.Fatal(errors.New("failed to verify volume get test-2"))
	}

	test3 := "PVC's of storage-volume\nExisting size of PVC: storage-volume-gg-timescaledb-0 is 150Gi\n\nPVC's of wal-volume\nExisting size of PVC: wal-volume-gg-timescaledb-0 is 20Gi\n\n"
	outputString = testVolumeGet(t, true, true, false)
	if outputString != test3 {
		t.Fatal(errors.New("failed to verify volume get test-3"))
	}

	test4 := "PVC's of storage-volume\nExisting size of PVC: storage-volume-gg-timescaledb-0 is 150Gi\n\n"
	outputString = testVolumeGet(t, true, false, false)
	if outputString != test4 {
		t.Fatal(errors.New("failed to verify volume get test-4"))
	}

	test5 := "PVC's of gg-prometheus-server\nExisting size of PVC: gg-prometheus-server is 8Gi\n\n"
	outputString = testVolumeGet(t, false, false, true)
	if outputString != test5 {
		t.Fatal(errors.New("failed to verify volume get test-5"))
	}

	// update default storageClass in Kind to allow pvc expansion
	err := k8s.UpdateStorageClassAllowVolumeExpand()
	if err != nil {
		t.Fatal(err)
	}

	testVolumeExpansion(t, "151Gi", "21Gi", "9Gi")
	res, err := k8s.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "151Gi" &&  res["wal-volume-gg-timescaledb-0"] != "21Gi" && res["gg-prometheus-server"] != "9Gi"{
		t.Fatal(errors.New("failed to verify volume expansion test-1"))
	}

	testVolumeExpansion(t, "152Gi", "22Gi", "")
	res, err = k8s.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "152Gi" &&  res["wal-volume-gg-timescaledb-0"] != "22Gi" && res["gg-prometheus-server"] != "9Gi"{
		t.Fatal(errors.New("failed to verify volume expansion test-2"))
	}

	testVolumeExpansion(t, "153Gi", "", "")
	res, err = k8s.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" &&  res["wal-volume-gg-timescaledb-0"] != "22Gi" && res["gg-prometheus-server"] != "9Gi"{
		t.Fatal(errors.New("failed to verify volume expansion test-3"))
	}

	testVolumeExpansion(t, "", "23Gi", "")
	res, err = k8s.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" &&  res["wal-volume-gg-timescaledb-0"] != "23Gi" && res["gg-prometheus-server"] != "9Gi"{
		t.Fatal(errors.New("failed to verify volume expansion test-4"))
	}

	testVolumeExpansion(t, "", "24Gi", "10Gi")
	res, err = k8s.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" &&  res["wal-volume-gg-timescaledb-0"] != "24Gi" && res["gg-prometheus-server"] != "10Gi"{
		t.Fatal(errors.New("failed to verify volume expansion test-5"))
	}

	testVolumeExpansion(t, "", "", "11Gi")
	res, err = k8s.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" &&  res["wal-volume-gg-timescaledb-0"] != "24Gi" && res["gg-prometheus-server"] != "11Gi"{
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}
}
