package installation_tests

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

func testDeleteData(t testing.TB, name, namespace string, k8sClient k8s.Client) {
	cmds := []string{"uninstall", "delete-data"}
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

	pvcs, err := k8sClient.KubeGetPVCNames("default", map[string]string{})
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

	// abc Install
	abcInstall := test_utils.TestInstallSpec{
		PathToChart:  PATH_TO_CHART,
		ReleaseName:  "abc",
		Namespace:    NAMESPACE,
		PathToValues: PATH_TO_TEST_VALUES,
		EnableBackUp: false,
		SkipWait:     true,
		OnlySecrets:  false,
	}
	abcInstall.TestInstall(t)

	// abc Uninstall
	abcUninstall := test_utils.TestUnInstallSpec{
		ReleaseName:  "abc",
		Namespace:    NAMESPACE,
		DeleteData:   false,
	}
	abcUninstall.TestUninstall(t)

	// def helm cmd install
	defInstall := test_utils.TestInstallSpec{
		PathToChart: PATH_TO_CHART,
		ReleaseName:  "def",
		Namespace:    NAMESPACE,
		PathToValues: PATH_TO_TEST_VALUES,
		EnableBackUp: false,
		SkipWait:     true,
		OnlySecrets:  false,
	}
	defInstall.TestInstall(t)

	// def uninstall
	defUninstall := test_utils.TestUnInstallSpec{
		ReleaseName: "def",
		Namespace:   NAMESPACE,
		DeleteData:  false,
	}
	defUninstall.TestUninstall(t)
	k8sClient := k8s.NewClient()
	testDeleteData(t, "def", "", k8sClient)
	time.Sleep(2 * time.Minute)
	// check pvc status post delete-data
	err := test_utils.CheckPVCSExist("def", "")
	if err != nil {
		t.Fatal(err)
	}

	// f1 install
	f1Install := test_utils.TestInstallSpec{
		PathToChart:  PATH_TO_CHART,
		ReleaseName:  "f1",
		Namespace:    NAMESPACE,
		PathToValues: "./../testdata/f1.yaml",
		EnableBackUp: false,
		SkipWait:     true,
		OnlySecrets:  false,
	}
	f1Install.TestInstall(t)

	// f1 uninstall
	f1Uninstall := test_utils.TestUnInstallSpec{
		ReleaseName: "f1",
		Namespace:   NAMESPACE,
		DeleteData:  false,
	}
	f1Uninstall.TestUninstall(t)

	// f2 install
	f2Install := test_utils.TestInstallSpec{
		PathToChart:  PATH_TO_CHART,
		ReleaseName:  "f2",
		Namespace:    NAMESPACE,
		PathToValues: "./../testdata/f2.yaml",
		EnableBackUp: false,
		SkipWait:     true,
		OnlySecrets:  false,
	}
	f2Install.TestInstall(t)
	// f2 uninstall
	f2Uninstall := test_utils.TestUnInstallSpec{
		ReleaseName: "f2",
		Namespace:   NAMESPACE,
		DeleteData:  false,
	}
	f2Uninstall.TestUninstall(t)

	// install --only-secrets
	f5Install := test_utils.TestInstallSpec{
		PathToChart:  PATH_TO_CHART,
		ReleaseName:  "f5",
		Namespace:    "secrets",
		PathToValues: PATH_TO_TEST_VALUES,
		EnableBackUp: false,
		SkipWait:     true,
		OnlySecrets:  true,
	}
	f5Install.TestInstall(t)
	pods, err := k8sClient.KubeGetAllPods("secrets", "f5")
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

	// kubectl get pods -A
	test_utils.ShowAllPods(t)

	dInstall := test_utils.TestInstallSpec{
		PathToChart:  PATH_TO_CHART,
		ReleaseName:  RELEASE_NAME,
		Namespace:    NAMESPACE,
		PathToValues: PATH_TO_TEST_VALUES,
		EnableBackUp: false,
		SkipWait:     false,
		OnlySecrets:  false,
	}
	dInstall.TestInstall(t)

	time.Sleep(3 * time.Minute)

	// test show values
	testHelmShowValues(t)

	// kubectl get pods -A
	test_utils.ShowAllPods(t)

	t.Logf("Waiting for pods to initialize...")
	pods, err = k8sClient.KubeGetAllPods(NAMESPACE, RELEASE_NAME)
	if err != nil {
		t.Logf("Error getting all pods")
		t.Fatal(err)
	}

	for _, pod := range pods {
		err = k8sClient.KubeWaitOnPod(NAMESPACE, pod.Name)
		if err != nil {
			t.Logf("Error while waiting on pod")
			t.Fatal(err)
		}
	}

	time.Sleep(30 * time.Second)
}
