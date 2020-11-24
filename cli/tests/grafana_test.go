package tests

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func testGrafanaPortForward(t testing.TB, port string) {
	cmds := []string{"grafana", "port-forward", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if port != "" {
		cmds = append(cmds, "-p", port)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command("./../bin/tobs", cmds...)

	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(4 * time.Second)

	if port == "" {
		port = "8080"
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

func testGrafanaGetPass(t testing.TB) {
	cmds := []string{"grafana", "get-password", "-n", RELEASE_NAME, "--namespace", NAMESPACE}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	getpass := exec.Command("./../bin/tobs", cmds...)

	out, err := getpass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testGrafanaChangePass(t testing.TB, newpass string) {
	cmds := []string{"grafana", "change-password", "\"" + newpass + "\"", "-n", RELEASE_NAME, "--namespace", NAMESPACE}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	changepass := exec.Command("./../bin/tobs", cmds...)

	out, err := changepass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func verifyGrafanaPass(t testing.TB, expectedPass string) {
	getpass := exec.Command("./../bin/tobs", "grafana", "get-password", "-n", RELEASE_NAME, "--namespace", NAMESPACE)

	out, err := getpass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	if string(out) == expectedPass {
		t.Fatalf("Password mismatch: got %v want %v", string(out), expectedPass)
	}
}

func TestGrafana(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Grafana tests")
	}

	testGrafanaPortForward(t, "")
	testGrafanaPortForward(t, "1235")
	testGrafanaPortForward(t, "2348")
	testGrafanaPortForward(t, "7390")

	testGrafanaGetPass(t)
	testGrafanaChangePass(t, "kraken")
	testGrafanaGetPass(t)
	verifyGrafanaPass(t, "kraken")
	testGrafanaChangePass(t, "cereal")
	testGrafanaGetPass(t)
	verifyGrafanaPass(t, "cereal")
	testGrafanaChangePass(t, "23498MSDF(*9389m*(@#M24309mDj")
	testGrafanaGetPass(t)
	verifyGrafanaPass(t, "23498MSDF(*9389m*(@#M24309mDj")
}
