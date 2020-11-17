package tests

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func testPromlensPortForward(t testing.TB, portPromlens, portConnector string) {
	cmds := []string{"promlens", "port-forward", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if portPromlens != "" {
		cmds = append(cmds, "-p", portPromlens)
	}
	if portConnector != "" {
		cmds = append(cmds, "-c", portConnector)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command("./../bin/tobs", cmds...)

	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	if portPromlens == "" {
		portPromlens = "8081"
	}

	if portConnector == "" {
		portConnector = "9201"
	}

	_, err = net.DialTimeout("tcp", "localhost:"+portPromlens, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	_, err = net.DialTimeout("tcp", "localhost:"+portConnector, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = portforward.Process.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPromlens(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Grafana tests")
	}

	testPromlensPortForward(t, "", "")
	testPromlensPortForward(t, "1235", "3421")
}
