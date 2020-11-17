package tests

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func testTimescaleGetPassword(t testing.TB, user string) {
	cmds := []string{"timescaledb", "get-password", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	getpass := exec.Command("./../bin/tobs", cmds...)

	out, err := getpass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testTimescaleChangePassword(t testing.TB, user, dbname, newpass string) {
	cmds := []string{"timescaledb", "change-password", newpass, "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	changepass := exec.Command("./../bin/tobs", cmds...)

	out, err := changepass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func verifyTimescalePassword(t testing.TB, user string, expectedPass string) {
	getpass := exec.Command("./../bin/tobs", "timescaledb", "get-password", "-U", user, "-n", RELEASE_NAME, "--namespace", NAMESPACE)

	out, err := getpass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	if string(out) == expectedPass {
		t.Fatalf("Password mismatch: got %v want %v", string(out), expectedPass)
	}
}

func testTimescalePortForward(t testing.TB, port string) {
	cmds := []string{"timescaledb", "port-forward", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
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

func testTimescaleConnect(t testing.TB, master bool, user string) {
	var connect *exec.Cmd

	if master {
		t.Logf("Running 'tobs timescaledb connect -m'")
		connect = exec.Command("./../bin/tobs", "timescaledb", "connect", "-m", "-n", RELEASE_NAME, "--namespace", NAMESPACE)
	} else {
		if user == "" {
			t.Logf("Running 'tobs timescaledb connect'")
			connect = exec.Command("./../bin/tobs", "timescaledb", "connect", "-n", RELEASE_NAME, "--namespace", NAMESPACE)
		} else {
			t.Logf("Running 'tobs timescaledb connect -U %v'", user)
			connect = exec.Command("./../bin/tobs", "timescaledb", "connect", "-U", user, "-n", RELEASE_NAME, "--namespace", NAMESPACE)
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

func TestTimescale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TimescaleDB tests")
	}

	testTimescaleGetPassword(t, "")
	testTimescaleChangePassword(t, "", "postgres", "battery")
	verifyTimescalePassword(t, "postgres", "battery")
	testTimescaleGetPassword(t, "admin")
	testTimescaleChangePassword(t, "admin", "", "chips")
	verifyTimescalePassword(t, "admin", "chips")

	testTimescalePortForward(t, "")
	testTimescalePortForward(t, "5432")
	testTimescalePortForward(t, "1789")
	testTimescalePortForward(t, "1030")
	testTimescalePortForward(t, "2389")

	testTimescaleConnect(t, true, "")
	testTimescaleConnect(t, false, "")
	testTimescaleConnect(t, false, "postgres")
	testTimescaleConnect(t, false, "postgres")
	testTimescaleConnect(t, false, "admin")
	testTimescaleConnect(t, false, "admin")
}
