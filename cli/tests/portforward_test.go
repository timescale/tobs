package tests

import (
	"net"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

func testpf(t testing.TB, timescale, grafana, prometheus, promscale, promlens string) {
	cmds := []string{"port-forward", "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if timescale != "" {
		cmds = append(cmds, "-t", timescale)
	}
	if grafana != "" {
		cmds = append(cmds, "-g", grafana)
	}
	if prometheus != "" {
		cmds = append(cmds, "-p", prometheus)
	}
	if promscale != "" {
		cmds = append(cmds, "-c", promscale)
	}
	if promlens != "" {
		cmds = append(cmds, "-l", promlens)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	portforward := exec.Command("./../bin/tobs", cmds...)

	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Second)

	if timescale == "" {
		timescale = "5432"
	}
	if grafana == "" {
		grafana = "8080"
	}
	if prometheus == "" {
		prometheus = "9090"
	}
	if promscale == "" {
		promscale = "9201"
	}
	if promlens == "" {
		promlens = "8081"
	}

	_, err = net.DialTimeout("tcp", "localhost:"+timescale, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.DialTimeout("tcp", "localhost:"+grafana, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.DialTimeout("tcp", "localhost:"+prometheus, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.DialTimeout("tcp", "localhost:"+promscale, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = net.DialTimeout("tcp", "localhost:"+promlens, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	err = portforward.Process.Signal(syscall.SIGINT)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPortForward(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping port-forwarding tests")
	}

	testpf(t, "", "", "", "", "")
	testpf(t, "3932", "", "", "", "")
	testpf(t, "", "4893", "", "", "")
	testpf(t, "", "", "2312", "", "")
	testpf(t, "4792", "4073", "", "", "")
	testpf(t, "", "5343", "9763", "", "")
	testpf(t, "9697", "6972", "", "", "")
	testpf(t, "1275", "4378", "1702", "", "")
	testpf(t, "4857", "2489", "3478", "", "")
	testpf(t, "4857", "2489", "3478", "8080", "4659")
}
