package tests

import (
    "fmt"
    "os/exec"
    "syscall"
    "testing"
    "time"
)

func testPrometheusPortForward(t testing.TB, port string) {
    var portforward *exec.Cmd

    if port == "" {
        fmt.Println("Running 'ts-obs prometheus port-forward'")
        portforward = exec.Command("/home/archen/go/bin/ts-obs", "prometheus", "port-forward")
    } else {
        fmt.Printf("Running 'ts-obs prometheus port-forward -p %v'\n", port)
        portforward = exec.Command("/home/archen/go/bin/ts-obs", "prometheus", "port-forward", "-p", port)
    }

    err := portforward.Start()
    if err != nil {
        t.Fatal(err)
    }
    time.Sleep(4 * time.Second)
    portforward.Process.Signal(syscall.SIGINT)
}

func TestPrometheus(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Prometheus tests")
    }

    testPrometheusPortForward(t, "")
    testPrometheusPortForward(t, "4")
    testPrometheusPortForward(t, "3920")
    testPrometheusPortForward(t, "7489")
    testPrometheusPortForward(t, "2238")
}
