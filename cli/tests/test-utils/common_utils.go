package test_utils

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

var PATH_TO_TOBS = "./../../bin/tobs"

type ReleaseInfo struct {
	Release   string
	Namespace string
}

func (r *ReleaseInfo) TestTimescaleGetPassword(t testing.TB, user string) {
	cmds := []string{"timescaledb", "get-password", "-n", r.Release, "--namespace", r.Namespace}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	getpass := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := getpass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func (r *ReleaseInfo) TestTimescaleChangePassword(t testing.TB, user, dbname, newpass string) {
	cmds := []string{"timescaledb", "change-password", newpass, "-n", r.Release, "--namespace", r.Namespace}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	changepass := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := changepass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func (r *ReleaseInfo) VerifyTimescalePassword(t testing.TB, user string, expectedPass string) {
	getpass := exec.Command(PATH_TO_TOBS, "timescaledb", "get-password", "-U", user, "-n", r.Release, "--namespace", r.Namespace)

	out, err := getpass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	if string(out) == expectedPass {
		t.Fatalf("Password mismatch: got %v want %v", string(out), expectedPass)
	}
}

func (r *ReleaseInfo) TestTimescalePortForward(t testing.TB, port string) {
	cmds := []string{"timescaledb", "port-forward", "-n", r.Release, "--namespace", r.Namespace}
	if port != "" {
		cmds = append(cmds, "-p", port)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command(PATH_TO_TOBS, cmds...)

	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(4 * time.Second)

	if port == "" {
		port = "5432"
	}

	_, err = net.DialTimeout("tcp", "localhost:"+port, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = portforward.Process.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatal(err)
	}
}

func (r *ReleaseInfo) TestTimescaleConnect(t testing.TB, master bool, user string) {
	var connect *exec.Cmd

	if master {
		t.Logf("Running 'tobs timescaledb connect -m'")
		connect = exec.Command(PATH_TO_TOBS, "timescaledb", "connect", "-m", "-n", r.Release, "--namespace", r.Namespace)
	} else {
		if user == "" {
			t.Logf("Running 'tobs timescaledb connect'")
			connect = exec.Command(PATH_TO_TOBS, "timescaledb", "connect", "-n", r.Release, "--namespace", r.Namespace)
		} else {
			t.Logf("Running 'tobs timescaledb connect -U %v'", user)
			connect = exec.Command(PATH_TO_TOBS, "timescaledb", "connect", "-U", user, "-n", r.Release, "--namespace", r.Namespace)
		}
	}

	err := connect.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Second)
	err = connect.Process.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatal(err)
	}
}

func (r *ReleaseInfo) TestPromscalePortForward(t testing.TB, portPromscale string) {
	cmds := []string{"promscale", "port-forward", "-n", r.Release, "--namespace", r.Namespace}

	if portPromscale != "" {
		cmds = append(cmds, "-p", portPromscale)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command(PATH_TO_TOBS, cmds...)

	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	if portPromscale == "" {
		portPromscale = "9201"
	}

	_, err = net.DialTimeout("tcp", "localhost:"+portPromscale, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = portforward.Process.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatal(err)
	}
}

func PortForwardPromscale(t testing.TB, releaseName, namespace string) {
	cmds := []string{"promscale", "port-forward", "-n", releaseName, "--namespace", namespace, "-p", "9201"}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command(PATH_TO_TOBS, cmds...)

	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	_, err = net.DialTimeout("tcp", "localhost:9201", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func (c *TestUnInstallSpec) TestUninstall(t testing.TB) {
	cmds := []string{"uninstall", "-n", c.ReleaseName, "--namespace", c.Namespace}
	if c.DeleteData {
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

type TestInstallSpec struct {
	PathToChart  string
	ReleaseName  string
	Namespace    string
	PathToValues string
	EnableBackUp bool
	SkipWait     bool
	OnlySecrets  bool
}

type TestUnInstallSpec struct {
	ReleaseName string
	Namespace   string
	DeleteData  bool
}

func (c *TestInstallSpec) TestInstall(t testing.TB) {
	cmds := []string{"install", "--chart-reference", c.PathToChart, "-n", c.ReleaseName, "--namespace", c.Namespace}
	if c.EnableBackUp {
		cmds = append(cmds, "--enable-timescaledb-backup")
	}
	if c.SkipWait {
		cmds = append(cmds, "--skip-wait")
	}
	if c.OnlySecrets {
		cmds = append(cmds, "--only-secrets")
	}
	if c.PathToValues != "" {
		cmds = append(cmds, "-f", c.PathToValues)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	install := exec.Command(PATH_TO_TOBS, cmds...)

	out, err := install.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func ShowAllPods(t testing.TB) {
	out := exec.Command("kubectl", "get", "pods", "-A")
	output, err := out.CombinedOutput()
	t.Log(string(output))
	if err != nil {
		t.Fatal(err)
	}
}

func ShowAllPVCs(t testing.TB) {
	out := exec.Command("kubectl", "get", "pvc", "-A")
	output, err := out.CombinedOutput()
	t.Log(string(output))
	if err != nil {
		t.Fatal(err)
	}
}
