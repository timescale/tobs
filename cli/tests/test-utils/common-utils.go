package test_utils

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

type ReleaseInfo struct {
	Release string
	Namespace string
}

func (r *ReleaseInfo) TestTimescaleGetPassword(t testing.TB, user string) {
	cmds := []string{"timescaledb", "get-password", "-n", r.Release, "--namespace", r.Namespace}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	getpass := exec.Command("./../../bin/tobs", cmds...)

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
	changepass := exec.Command("./../../bin/tobs", cmds...)

	out, err := changepass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func (r *ReleaseInfo) VerifyTimescalePassword(t testing.TB, user string, expectedPass string) {
	getpass := exec.Command("./../../bin/tobs", "timescaledb", "get-password", "-U", user, "-n", r.Release, "--namespace", r.Namespace)

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
	portforward := exec.Command("./../../bin/tobs", cmds...)

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
		connect = exec.Command("./../../bin/tobs", "timescaledb", "connect", "-m", "-n", r.Release, "--namespace", r.Namespace)
	} else {
		if user == "" {
			t.Logf("Running 'tobs timescaledb connect'")
			connect = exec.Command("./../../bin/tobs", "timescaledb", "connect", "-n", r.Release, "--namespace", r.Namespace)
		} else {
			t.Logf("Running 'tobs timescaledb connect -U %v'", user)
			connect = exec.Command("./../../bin/tobs", "timescaledb", "connect", "-U", user, "-n", r.Release, "--namespace", r.Namespace)
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

func (r *ReleaseInfo)  TestPromscalePortForward(t testing.TB, portPromscale string) {
	cmds := []string{"promscale", "port-forward", "-n", r.Release, "--namespace", r.Namespace}

	if portPromscale != "" {
		cmds = append(cmds, "-p", portPromscale)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command("./../../bin/tobs", cmds...)

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