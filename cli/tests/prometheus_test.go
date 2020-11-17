package tests

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func testPrometheusPortForward(t testing.TB, port string) {
	cmds := []string{"prometheus", "port-forward", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
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
		port = "9090"
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
