package tests

import (
	"net"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func testTimescaleGetPassword(t testing.TB, user string) {
	var getpass *exec.Cmd

	if user == "" {
		t.Logf("Running 'ts-obs timescaledb get-password'")
		getpass = exec.Command("ts-obs", "timescaledb", "get-password")
	} else {
		t.Logf("Running 'ts-obs timescaledb get-password -U %v'\n", user)
		getpass = exec.Command("ts-obs", "timescaledb", "get-password", "-U", user)
	}

	out, err := getpass.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testTimescalePortForward(t testing.TB, port string) {
	var portforward *exec.Cmd

	if port == "" {
		t.Logf("Running 'ts-obs timescaledb port-forward'")
		portforward = exec.Command("ts-obs", "timescaledb", "port-forward")
	} else {
		t.Logf("Running 'ts-obs timescaledb port-forward -p %v'\n", port)
		portforward = exec.Command("ts-obs", "timescaledb", "port-forward", "-p", port)
	}

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

	portforward.Process.Signal(syscall.SIGINT)

}

func testTimescaleConnect(t testing.TB, master bool, user, password string) {
	var connect *exec.Cmd

	if master {
		t.Logf("Running 'ts-obs timescaledb connect -m'")
		connect = exec.Command("ts-obs", "timescaledb", "connect", "-m")
	} else {
		if user == "" {
			t.Logf("Running 'ts-obs timescaledb connect -p %v'\n", password)
			connect = exec.Command("ts-obs", "timescaledb", "connect", "-p", password)
		} else {
			t.Logf("Running 'ts-obs timescaledb connect -U %v -p %v'\n", user, password)
			connect = exec.Command("ts-obs", "timescaledb", "connect", "-U", user, "-p", password)
		}
	}

	err := connect.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Second)
	connect.Process.Signal(syscall.SIGINT)

	deletepod := exec.Command("kubectl", "delete", "pods", "psql")
	deletepod.Run()
}

func TestTimescale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TimescaleDB tests")
	}

	testTimescaleGetPassword(t, "")
	testTimescaleGetPassword(t, "admin")
	testTimescaleGetPassword(t, "23948")
	testTimescaleGetPassword(t, "user93")

	testTimescalePortForward(t, "")
	testTimescalePortForward(t, "5432")
	testTimescalePortForward(t, "1789")
	testTimescalePortForward(t, "1030")
	testTimescalePortForward(t, "2389")

	os.Setenv("TESTP", "tea")
	os.Setenv("TESTQ", "cola")
	os.Setenv("TESTR", "mug")

	testTimescaleConnect(t, true, "", "")
	testTimescaleConnect(t, false, "", "TESTP")
	testTimescaleConnect(t, false, "postgres", "TESTP")
	testTimescaleConnect(t, false, "postgres", "TESTR")
	testTimescaleConnect(t, false, "admin", "TESTP")
	testTimescaleConnect(t, false, "admin", "TESTQ")
}
