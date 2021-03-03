package tobs_cli_tests

import (
	"errors"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"github.com/timescale/tobs/cli/tests/test-utils"
	v1 "k8s.io/api/core/v1"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func testVolumeExpansion(t testing.TB, timescaleDBStorage, timescaleDBWal, prometheusStorage string, restartPods bool) {
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

	if restartPods {
		cmds = append(cmds, "--restart-pods")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	expand := exec.Command(PATH_TO_TOBS, cmds...)
	_, err := expand.CombinedOutput()
	if err != nil {
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
	expand := exec.Command(PATH_TO_TOBS, cmds...)
	out, err := expand.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	return string(out)
}

func TestVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Volume tests")
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
	err := test_utils.UpdateStorageClassAllowVolumeExpand()
	if err != nil {
		t.Fatal(err)
	}

	testVolumeExpansion(t, "151Gi", "21Gi", "9Gi", false)
	res, err := test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "151Gi" && res["wal-volume-gg-timescaledb-0"] != "21Gi" && res["gg-prometheus-server"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-1"))
	}

	testVolumeExpansion(t, "152Gi", "22Gi", "", false)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "152Gi" && res["wal-volume-gg-timescaledb-0"] != "22Gi" && res["gg-prometheus-server"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-2"))
	}

	testVolumeExpansion(t, "153Gi", "", "", false)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" && res["wal-volume-gg-timescaledb-0"] != "22Gi" && res["gg-prometheus-server"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-3"))
	}

	testVolumeExpansion(t, "", "23Gi", "", false)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" && res["wal-volume-gg-timescaledb-0"] != "23Gi" && res["gg-prometheus-server"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-4"))
	}

	testVolumeExpansion(t, "", "24Gi", "10Gi", false)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" && res["wal-volume-gg-timescaledb-0"] != "24Gi" && res["gg-prometheus-server"] != "10Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-5"))
	}

	testVolumeExpansion(t, "", "", "11Gi", false)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" && res["wal-volume-gg-timescaledb-0"] != "24Gi" && res["gg-prometheus-server"] != "11Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	timescaleDBLabels := map[string]string{"app": "gg-timescaledb", "release": "gg"}
	prometheusLabels := map[string]string{"app": "prometheus", "component": "server", "release": "gg"}

	// TESTCASE: Volume expand Prometheus storage and restart the pods

	pods, err := k8s.KubeGetPods("ns", prometheusLabels)
	if err != nil {
		t.Fatal(err)
	}

	var podsSet []podDetails
	for _, p := range pods {
		if p.Status.Phase == "Running" {
			details := podDetails{
				name:                  p.Name,
				oldPodCreateTimestamp: p.CreationTimestamp.String(),
			}
			podsSet = append(podsSet, details)
		}
	}

	testVolumeExpansion(t, "", "", "12Gi", true)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "153Gi" && res["wal-volume-gg-timescaledb-0"] != "24Gi" && res["gg-prometheus-server"] != "12Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	time.Sleep(8 * time.Second)
	var labels []map[string]string
	labels = append(labels, timescaleDBLabels)
	verifyPodRestart(t, podsSet, labels)

	// TESTCASE: Volume expand TimescaleDB storage, wal, Prometheus storage & restart the pods

	// sleep between test executions
	time.Sleep(30 * time.Second)
	pods, err = k8s.KubeGetPods("ns", prometheusLabels)
	if err != nil {
		t.Fatal(err)
	}

	podsSet = []podDetails{}
	labels = []map[string]string{}
	for _, p := range pods {
		if p.Status.Phase == "Running" {
			details := podDetails{
				name:                  p.Name,
				oldPodCreateTimestamp: p.CreationTimestamp.String(),
			}
			podsSet = append(podsSet, details)
		}
	}

	pods, err = k8s.KubeGetPods("ns", timescaleDBLabels)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range pods {
		if p.Status.Phase == "Running" {
			details := podDetails{
				name:                  p.Name,
				oldPodCreateTimestamp: p.CreationTimestamp.String(),
			}
			podsSet = append(podsSet, details)
		}
	}

	testVolumeExpansion(t, "154Gi", "24Gi", "13Gi", true)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "154Gi" && res["wal-volume-gg-timescaledb-0"] != "24Gi" && res["gg-prometheus-server"] != "13Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	// As timescaleDB pod is an statefulset bounded to a PVC
	// it takes some time to move to terminating state
	time.Sleep(1 * time.Minute)

	labels = append(labels, timescaleDBLabels, prometheusLabels)
	verifyPodRestart(t, podsSet, labels)

	// TESTCASE: Volume Expand only timescaleDB and restart pods.

	// sleep between test executions
	time.Sleep(30 * time.Second)
	pods, err = k8s.KubeGetPods("ns", map[string]string{"app": "gg-timescaledb", "release": "gg"})
	if err != nil {
		t.Fatal(err)
	}

	podsSet = []podDetails{}
	labels = []map[string]string{}
	for _, p := range pods {
		if p.Status.Phase == "Running" {
			details := podDetails{
				name:                  p.Name,
				oldPodCreateTimestamp: p.CreationTimestamp.String(),
			}
			podsSet = append(podsSet, details)
		}
	}

	testVolumeExpansion(t, "155Gi", "25Gi", "", true)
	res, err = test_utils.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-gg-timescaledb-0"] != "155Gi" && res["wal-volume-gg-timescaledb-0"] != "25Gi" && res["gg-prometheus-server"] != "13Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	// As timescaleDB pod is an statefulset bounded to a PVC
	// it takes some time to move to terminating state
	time.Sleep(1 * time.Minute)

	labels = append(labels, timescaleDBLabels)
	verifyPodRestart(t, podsSet, labels)

}

func verifyPodRestart(t *testing.T, podsSet []podDetails, labels []map[string]string) {
	var pods []v1.Pod

	// sleep for 2 mins as Prometheus & TimescaleDB termination takes
	// approx 2 mins for restart to happen.
	time.Sleep(2 * time.Minute)

	for _, l := range labels {
		p, err := k8s.KubeGetPods("ns", l)
		if err != nil {
			t.Fatal(err)
		}
		pods = append(pods, p...)
	}

	podsCounter := 0
	for _, p := range pods {
		if p.Status.Phase == "Running" {
			podsCounter++
			for _, v := range podsSet {
				// compare names & timestamp as statefulsets will have same name then compare create timestamp
				if v.name == p.Name && v.oldPodCreateTimestamp == p.CreationTimestamp.String() {
					t.Fatal(errors.New("failed to restart the pod after volume expansion " + v.name))
				}
			}
		}
	}

	// validate number of pods killed are equal to number of pods running i.e expected to be restarted.
	if len(podsSet) != podsCounter {
		t.Fatal(errors.New("failed to restart pods after volume expansion"))
	}
}

type podDetails struct {
	name                  string
	oldPodCreateTimestamp string
}
