package tests

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"
)

func testInstall(t testing.TB, name, namespace, filename string) {
	cmds := []string{"install", "--chart-reference", "../../chart"}
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

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	install := exec.Command("./../bin/tobs", cmds...)

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testHelmInstall(t testing.TB, name, namespace, filename string) {
	cmds := []string{"helm", "install", "--chart-reference", "../../chart"}
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

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	install := exec.Command("./../bin/tobs", cmds...)

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
	uninstall := exec.Command("./../bin/tobs", cmds...)

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
	uninstall := exec.Command("./../bin/tobs", cmds...)

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
	deletedata := exec.Command("./../bin/tobs", cmds...)

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
	showvalues = exec.Command("./../bin/tobs", "helm", "show-values")

	out, err := showvalues.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func TestInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping installation tests")
	}

	testHelmShowValues(t)

	testUninstall(t, "", "", false)
	testInstall(t, "", "", "")
	testHelmUninstall(t, "", "", true)
	testHelmInstall(t, "", "", "")
	testUninstall(t, "", "", false)
	testHelmDeleteData(t, "", "")
	testHelmInstall(t, "", "", "")
	testHelmUninstall(t, "", "", false)
	testInstall(t, "", "", "")
	testUninstall(t, "", "", false)
	testHelmDeleteData(t, "", "")

	testInstall(t, "sd-fo9ods-oe93", "", "")
	testHelmUninstall(t, "sd-fo9ods-oe93", "", false)
	testHelmInstall(t, "x98-2cn4-ru2-9cn48u", "nondef", "")
	testUninstall(t, "x98-2cn4-ru2-9cn48u", "nondef", false)
	testHelmInstall(t, "as-dn-in234i-n", "", "")
	testHelmUninstall(t, "as-dn-in234i-n", "", false)
	testInstall(t, "we-3oiwo3o-s-d", "", "")
	testUninstall(t, "we-3oiwo3o-s-d", "", false)

	testInstall(t, "f1", "", "./testdata/f1.yml")
	testHelmUninstall(t, "f1", "", false)
	testHelmInstall(t, "f2", "", "./testdata/f2.yml")
	testUninstall(t, "f2", "", false)
	testHelmInstall(t, "f3", "nas", "./testdata/f3.yml")
	testHelmUninstall(t, "f3", "nas", false)
	testInstall(t, "f4", "", "./testdata/f4.yml")
	testUninstall(t, "f4", "", false)

	testInstall(t, "", "", "")

	time.Sleep(10 * time.Second)

	t.Logf("Waiting for pods to initialize...")
	pods, err := k8s.KubeGetAllPods(NAMESPACE, RELEASE_NAME)
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
