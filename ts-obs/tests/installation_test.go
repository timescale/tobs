package tests

import (
	"os/exec"
	"testing"
	"time"

	"ts-obs/cmd"
)

func testInstall(t testing.TB, name string, filename string) {
	var install *exec.Cmd
	if name == "" && filename == "" {
		t.Logf("Running 'ts-obs install'")
		install = exec.Command("ts-obs", "install")
	} else if name == "" {
		t.Logf("Running 'ts-obs install -f %v'\n", filename)
		install = exec.Command("ts-obs", "install", "-f", filename)
	} else if filename == "" {
		t.Logf("Running 'ts-obs install -n %v'\n", name)
		install = exec.Command("ts-obs", "install", "-n", name)
	} else {
		t.Logf("Running 'ts-obs install -n %v -f %v'\n", name, filename)
		install = exec.Command("ts-obs", "install", "-n", name, "-f", filename)
	}

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testHelmInstall(t testing.TB, name string, filename string) {
	var install *exec.Cmd

	if name == "" && filename == "" {
		t.Logf("Running 'ts-obs helm install'")
		install = exec.Command("ts-obs", "helm", "install")
	} else if name == "" {
		t.Logf("Running 'ts-obs helm install -f %v'\n", filename)
		install = exec.Command("ts-obs", "helm", "install", "-f", filename)
	} else if filename == "" {
		t.Logf("Running 'ts-obs helm install -n %v'\n", name)
		install = exec.Command("ts-obs", "helm", "install", "-n", name)
	} else {
		t.Logf("Running 'ts-obs helm install -n %v -f %v'\n", name, filename)
		install = exec.Command("ts-obs", "helm", "install", "-n", name, "-f", filename)
	}

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testUninstall(t testing.TB, name string) {
	var uninstall *exec.Cmd

	if name == "" {
		t.Logf("Running 'ts-obs uninstall'")
		uninstall = exec.Command("ts-obs", "uninstall")
	} else {
		t.Logf("Running 'ts-obs uninstall -n %v'\n", name)
		uninstall = exec.Command("ts-obs", "uninstall", "-n", name)
	}

	out, err := uninstall.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testHelmUninstall(t testing.TB, name string) {
	var uninstall *exec.Cmd

	if name == "" {
		t.Logf("Running 'ts-obs helm uninstall'")
		uninstall = exec.Command("ts-obs", "uninstall")
	} else {
		t.Logf("Running 'ts-obs helm uninstall -n %v'\n", name)
		uninstall = exec.Command("ts-obs", "uninstall", "-n", name)
	}

	out, err := uninstall.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	pods, err := cmd.KubeGetAllPods("ts-obs", "default")
	if len(pods) != 0 {
		t.Fatal("Pod remaining after uninstall")
	}

}

func testHelmGetYaml(t testing.TB) {
	var getyaml *exec.Cmd

	t.Logf("Running 'ts-obs helm get-yaml'")
	getyaml = exec.Command("ts-obs", "helm", "get-yaml")

	out, err := getyaml.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func TestInstallation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping installation tests")
	}

	testHelmGetYaml(t)

	testUninstall(t, "")
	testInstall(t, "", "")
	testHelmUninstall(t, "")
	testHelmInstall(t, "", "")
	testUninstall(t, "")
	testHelmInstall(t, "", "")
	testHelmUninstall(t, "")
	testInstall(t, "", "")
	testUninstall(t, "")

	testInstall(t, "sd-fo9ods-oe93", "")
	testHelmUninstall(t, "sd-fo9ods-oe93")
	testHelmInstall(t, "x98-2cn4-ru2-9cn48u", "")
	testUninstall(t, "x98-2cn4-ru2-9cn48u")
	testHelmInstall(t, "as-dn-in234i-n", "")
	testHelmUninstall(t, "as-dn-in234i-n")
	testInstall(t, "we-3oiwo3o-s-d", "")
	testUninstall(t, "we-3oiwo3o-s-d")

	testInstall(t, "f1", "./testdata/f1.yml")
	testHelmUninstall(t, "f1")
	testHelmInstall(t, "f2", "./testdata/f2.yml")
	testUninstall(t, "f2")
	testHelmInstall(t, "f3", "./testdata/f3.yml")
	testHelmUninstall(t, "f3")
	testInstall(t, "f4", "./testdata/f4.yml")
	testUninstall(t, "f4")

	testInstall(t, "", "")

	time.Sleep(10 * time.Second)

	t.Logf("Waiting for pods to initialize...")
	pods, err := cmd.KubeGetAllPods("ts-obs", "default")
	if err != nil {
		t.Logf("Error getting all pods")
		t.Fatal(err)
	}

	for _, pod := range pods {
		err = cmd.KubeWaitOnPod("default", pod.Name)
		if err != nil {
			t.Logf("Error while waiting on pod")
			t.Fatal(err)
		}
	}
}
