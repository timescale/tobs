package tests

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func testPromscalePortForward(t testing.TB, portPromscale string) {
	cmds := []string{"promscale", "port-forward", "-n", RELEASE_NAME, "--namespace", NAMESPACE}

	if portPromscale != "" {
		cmds = append(cmds, "-p", portPromscale)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command("./../bin/tobs", cmds...)

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

func TestPromscale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Promscale tests")
	}

	testPromscalePortForward(t, "")
	testPromscalePortForward(t, "3421")
}

