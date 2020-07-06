package tests

import (
	"net"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func testPrometheusPortForward(t testing.TB, port string) {
	var portforward *exec.Cmd

	if port == "" {
		t.Logf("Running 'ts-obs prometheus port-forward'")
		portforward = exec.Command("ts-obs", "prometheus", "port-forward", "-n", RELEASE_NAME, "--namespace", NAMESPACE)
	} else {
		t.Logf("Running 'ts-obs prometheus port-forward -p %v'", port)
		portforward = exec.Command("ts-obs", "prometheus", "port-forward", "-p", port, "-n", RELEASE_NAME, "--namespace", NAMESPACE)
	}

	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(4 * time.Second)

	if port == "" {
		port = "9090"
	}

	_, err = net.DialTimeout("tcp", "localhost:"+port, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	portforward.Process.Signal(syscall.SIGINT)
}

func TestPrometheus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Prometheus tests")
	}

	testPrometheusPortForward(t, "")
	testPrometheusPortForward(t, "2398")
	testPrometheusPortForward(t, "3920")
	testPrometheusPortForward(t, "7489")
	testPrometheusPortForward(t, "2238")
}
