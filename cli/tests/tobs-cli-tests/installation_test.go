package tobs_cli_tests

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

func testInstall(t testing.TB, name, namespace, filename string, enableBackUp, skipWait, onlySecrets bool) {
	cmds := []string{"install", "--chart-reference", PATH_TO_CHART}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if filename != "" {
		cmds = append(cmds, "-f", filename)
	}
	if enableBackUp {
		cmds = append(cmds, "--enable-timescaledb-backup")
	}
	if skipWait {
		cmds = append(cmds, "--skip-wait")
	}
	if onlySecrets {
		cmds = append(cmds, "--only-secrets")
	}
	if filename == "" {
		filename = PATH_TO_TEST_VALUES
	}
	cmds = append(cmds, "-f", filename)

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	install := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testHelmInstall(t testing.TB, name, namespace, filename string) {
	cmds := []string{"helm", "install", "--chart-reference", PATH_TO_CHART}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if filename == "" {
		filename = PATH_TO_TEST_VALUES
	}
	cmds = append(cmds, "-f", filename)

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	install := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testUninstall(t testing.TB, name, namespace string, deleteData bool) {
	cmds := []string{"uninstall"}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if deleteData {
		cmds = append(cmds, "--delete-data")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	uninstall := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := uninstall.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testHelmUninstall(t testing.TB, name, namespace string, deleteData bool) {
	cmds := []string{"helm", "uninstall"}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}
	if deleteData {
		cmds = append(cmds, "--delete-data")
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	uninstall := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := uninstall.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	pods, err := k8s.KubeGetAllPods("tobs", "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(pods) != 0 {
		t.Fatal("Pod remaining after uninstall")
	}

}

func testHelmDeleteData(t testing.TB, name, namespace string) {
	cmds := []string{"helm", "delete-data"}
	if name != "" {
		cmds = append(cmds, "-n", name)
	} else {
		cmds = append(cmds, "-n", RELEASE_NAME)
	}
	if namespace != "" {
		cmds = append(cmds, "--namespace", namespace)
	} else {
		cmds = append(cmds, "--namespace", NAMESPACE)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	deletedata := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := deletedata.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	pvcs, err := k8s.KubeGetPVCNames("default", map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(pvcs) != 0 {
		t.Fatal("PVC remaining")
	}
}

func testHelmShowValues(t testing.TB) {
	var showvalues *exec.Cmd

	t.Logf("Running 'tobs helm show-values'")

	showvalues = exec.Command(PATH_TO_TOBS, "helm", "show-values", "-c", PATH_TO_CHART)
	out, err := showvalues.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func TestInstallation(t *testing.T) {
	v := os.Getenv("SKIP_INSTALL_TESTS")

	if testing.Short() || v == "TRUE" {
		t.Skip("Skipping installation tests")
	}

	testHelmShowValues(t)

	testUninstall(t, "", "", true)

	testInstall(t, "abc", "", "", false, true, false)
	testHelmUninstall(t, "abc", "", false)

	testHelmInstall(t, "def", "", "")
	testUninstall(t, "def", "", false)
	testHelmDeleteData(t, "def", "")

	testInstall(t, "f1", "", "./../testdata/f1.yaml", false, true, false)
	testHelmUninstall(t, "f1", "", false)

	testHelmInstall(t, "f2", "", "./../testdata/f2.yaml")
	testUninstall(t, "f2", "", false)

	// install --only-secrets
	testInstall(t, "f5", "secrets", "", false, false, true)
	pods, err := k8s.KubeGetAllPods("secrets", "f5")
	if err != nil {
		t.Log("failed to get all tobs pods")
		t.Fatal(err)
	}
	if len(pods) != 0 {
		t.Fatal("failed to install tobs with --only-secrets. We see other pods by tobs install")
	}
	err = test_utils.DeleteNamespace("secrets")
	if err != nil {
		t.Fatal(err)
	}

	// This installation is used to run all tests in tobs-cli-tests
	testInstall(t, "", "", "", false, false, false)

	time.Sleep(2 * time.Minute)

	t.Logf("Waiting for pods to initialize...")
	pods, err = k8s.KubeGetAllPods(NAMESPACE, RELEASE_NAME)
	if err != nil {
		t.Logf("Error getting all pods")
		t.Fatal(err)
	}

	for _, pod := range pods {
		err = k8s.KubeWaitOnPod(NAMESPACE, pod.Name)
		if err != nil {
			t.Logf("Error while waiting on pod")
			t.Fatal(err)
		}
	}

	time.Sleep(30 * time.Second)
}
