package tests

import (
    "fmt"
    "os/exec"
    "syscall"
    "testing"
    "time"
)

func testpf(t testing.TB, timescale, grafana, prometheus string) {
    var portforward *exec.Cmd

    if timescale == "" {
        if grafana == "" {
            if prometheus == "" {
                fmt.Printf("Running 'ts-obs port-forward'\n")
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward")
            } else {
                fmt.Printf("Running 'ts-obs port-forward -p %v'\n", prometheus)
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward", "-p", prometheus)
            }
        } else {
            if prometheus == "" {
                fmt.Printf("Running 'ts-obs port-forward -g %v'\n", grafana)
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward", "-g", grafana)
            } else {
                fmt.Printf("Running 'ts-obs port-forward -g %v -p %v'\n", grafana, prometheus)
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward", "-g", grafana, "-p", prometheus)
            }
        }
    } else {
        if grafana == "" {
            if prometheus == "" {
                fmt.Printf("Running 'ts-obs port-forward -t %v'\n", timescale)
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward", "-t", timescale)
            } else {
                fmt.Printf("Running 'ts-obs port-forward -t %v -p %v'\n", timescale, prometheus)
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward", "-t", timescale, "-p", prometheus)
            }
        } else {
            if prometheus == "" {
                fmt.Printf("Running 'ts-obs port-forward -t %v -g %v'\n", timescale, grafana)
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward", "-t", timescale, "-g", grafana)
            } else {
                fmt.Printf("Running 'ts-obs port-forward -t %v -g %v -p %v'\n", timescale, grafana, prometheus)
                portforward = exec.Command("/home/archen/go/bin/ts-obs", "port-forward", "-t", timescale, "-g", grafana, "-p", prometheus)
            }
        }
    }

    err := portforward.Start()
    if err != nil {
        t.Fatal(err)
    }
    time.Sleep(4 * time.Second)
    portforward.Process.Signal(syscall.SIGINT)
}

func TestPortForward(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping port-forwarding tests")
    }

    testpf(t, "", "", "")
    testpf(t, "3932", "", "")
    testpf(t, "", "4893", "")
    testpf(t, "", "", "2312")
    testpf(t, "4792", "4073", "")
    testpf(t, "", "5343", "9763")
    testpf(t, "9697", "6972", "")
    testpf(t, "1275", "4378", "1702")
    testpf(t, "4857", "2489", "3478")
}
