package tobs_cli_tests

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
	v1 "k8s.io/api/core/v1"
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
		cmds = append(cmds, "--restart-pods", "--force-kill")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	expand := exec.Command(PATH_TO_TOBS, cmds...)
	output, err := expand.CombinedOutput()
	if err != nil {
		t.Log(string(output))
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
		t.Log(string(out))
		// kubectl get pods -A
		out := exec.Command("kubectl", "get", "pods", "-A")
		output, err := out.CombinedOutput()
		if err != nil {
			t.Log(string(output))
			t.Log(err)
		}
		t.Log(string(output))
		t.Fatal(err)
	}

	return string(out)
}

func TestVolume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Volume tests")
	}

	test1 := "PVC's of storage-volume\nExisting size of PVC: storage-volume-" + RELEASE_NAME + "-timescaledb-0 is 150Gi\n\nPVC's of wal-volume\nExisting size of PVC: wal-volume-" + RELEASE_NAME + "-timescaledb-0 is 20Gi\n\nPVC's of prometheus-tobs-kube-prometheus-prometheus-db\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0 is 8Gi\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1 is 8Gi\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2 is 8Gi\n\n"
	outputString := testVolumeGet(t, true, true, true)
	if outputString != test1 {
		t.Log("expected: ", test1)
		t.Log("result: ", outputString)

		// kubectl get pods -A
		test_utils.ShowAllPods(t)

		// kubectl get pvc -A
		test_utils.ShowAllPVCs(t)

		t.Fatal(errors.New("failed to verify volume get test-1"))
	}

	test2 := "PVC's of wal-volume\nExisting size of PVC: wal-volume-" + RELEASE_NAME + "-timescaledb-0 is 20Gi\n\nPVC's of prometheus-tobs-kube-prometheus-prometheus-db\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0 is 8Gi\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1 is 8Gi\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2 is 8Gi\n\n"
	outputString = testVolumeGet(t, false, true, true)
	if outputString != test2 {
		t.Log("expected: ", test2)
		t.Log("result: ", outputString)
		t.Fatal(errors.New("failed to verify volume get test-2"))
	}

	test3 := "PVC's of storage-volume\nExisting size of PVC: storage-volume-" + RELEASE_NAME + "-timescaledb-0 is 150Gi\n\nPVC's of wal-volume\nExisting size of PVC: wal-volume-" + RELEASE_NAME + "-timescaledb-0 is 20Gi\n\n"
	outputString = testVolumeGet(t, true, true, false)
	if outputString != test3 {
		t.Log("expected: ", test3)
		t.Log("result: ", outputString)
		t.Fatal(errors.New("failed to verify volume get test-3"))
	}

	test4 := "PVC's of storage-volume\nExisting size of PVC: storage-volume-" + RELEASE_NAME + "-timescaledb-0 is 150Gi\n\n"
	outputString = testVolumeGet(t, true, false, false)
	if outputString != test4 {
		t.Log("expected: ", test4)
		t.Log("result: ", outputString)
		t.Fatal(errors.New("failed to verify volume get test-4"))
	}

	test5 := "PVC's of prometheus-tobs-kube-prometheus-prometheus-db\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0 is 8Gi\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1 is 8Gi\nExisting size of PVC: prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2 is 8Gi\n\n"
	outputString = testVolumeGet(t, false, false, true)
	if outputString != test5 {
		t.Log("expected: ", test5)
		t.Log("result: ", outputString)
		t.Fatal(errors.New("failed to verify volume get test-5"))
	}

	// update default storageClass in Kind to allow pvc expansion
	err := kubeClient.UpdateStorageClassAllowVolumeExpand()
	if err != nil {
		t.Fatal(err)
	}

	testVolumeExpansion(t, "151Gi", "21Gi", "9Gi", false)
	res, err := kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "151Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "21Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-1"))
	}

	testVolumeExpansion(t, "152Gi", "22Gi", "", false)
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "152Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "22Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-2"))
	}

	testVolumeExpansion(t, "153Gi", "", "", false)
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "153Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "22Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-3"))
	}

	testVolumeExpansion(t, "", "23Gi", "", false)
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "153Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "23Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "9Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "9Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-4"))
	}

	testVolumeExpansion(t, "", "24Gi", "10Gi", false)
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "153Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "24Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "10Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "10Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "10Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-5"))
	}

	testVolumeExpansion(t, "", "", "11Gi", false)
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "153Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "24Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "11Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "11Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "11Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	timescaleDBLabels := map[string]string{"app": RELEASE_NAME + "-timescaledb", "release": RELEASE_NAME}
	prometheusLabels := map[string]string{"app": "prometheus", "prometheus": "tobs-kube-prometheus-prometheus"}

	// TESTCASE: Volume expand Prometheus storage and restart the pods

	pods, err := kubeClient.K8s.KubeGetPods(NAMESPACE, prometheusLabels)
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
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "153Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "24Gi" &&
		res["prometheus-tobs-kube-prometheus-tobs-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "12Gi" &&
		res["prometheus-tobs-kube-prometheus-tobs-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "12Gi" &&
		res["prometheus-tobs-kube-prometheus-tobs-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "12Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	var labels []map[string]string
	labels = append(labels, prometheusLabels)
	// pod restart takes sometime for Prometheus to start & to move to running state.
	time.Sleep(1 * time.Minute)
	verifyPodRestart(t, podsSet, labels)

	// TESTCASE: Volume expand TimescaleDB storage, wal, Prometheus storage & restart the pods

	// sleep between test executions
	time.Sleep(30 * time.Second)
	pods, err = kubeClient.K8s.KubeGetPods(NAMESPACE, prometheusLabels)
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

	pods, err = kubeClient.K8s.KubeGetPods(NAMESPACE, timescaleDBLabels)
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
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "154Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "24Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "13Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "13Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "13Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	labels = append(labels, timescaleDBLabels, prometheusLabels)
	// pod restart takes sometime for TimescaleDB to start & to move to running state.
	time.Sleep(1 * time.Minute)
	verifyPodRestart(t, podsSet, labels)

	// TESTCASE: Volume Expand only timescaleDB and restart pods.

	// sleep between test executions
	time.Sleep(30 * time.Second)
	pods, err = kubeClient.K8s.KubeGetPods(NAMESPACE, timescaleDBLabels)
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
	res, err = kubeClient.GetAllPVCSizes()
	if err != nil {
		t.Fatal(err)
	}
	if res["storage-volume-"+RELEASE_NAME+"-timescaledb-0"] != "155Gi" && res["wal-volume-"+RELEASE_NAME+"-timescaledb-0"] != "25Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-0"] != "13Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-1"] != "13Gi" &&
		res["prometheus-tobs-kube-prometheus-prometheus-db-prometheus-tobs-kube-prometheus-prometheus-2"] != "13Gi" {
		t.Fatal(errors.New("failed to verify volume expansion test-6"))
	}

	labels = append(labels, timescaleDBLabels)
	// pod restart takes sometime for TimescaleDB to start & to move to running state.
	time.Sleep(1 * time.Minute)
	verifyPodRestart(t, podsSet, labels)
}

func verifyPodRestart(t *testing.T, podsSet []podDetails, labels []map[string]string) {
	var pods []v1.Pod

	for _, l := range labels {
		p, err := kubeClient.K8s.KubeGetPods(NAMESPACE, l)
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
					t.Log(p)
					test_utils.ShowAllPods(t)
					t.Fatal(errors.New("failed to restart the pod after volume expansion " + v.name))
				}
			}
		}
	}

	// validate number of pods killed are equal to number of pods running i.e expected to be restarted.
	if len(podsSet) != podsCounter {
		test_utils.ShowAllPods(t)
		t.Logf("Number of Pods %v, PodsCounter %v", len(podsSet), podsCounter)
		t.Fatal(errors.New("failed to restart pods after volume expansion"))
	}
}

type podDetails struct {
	name                  string
	oldPodCreateTimestamp string
}
