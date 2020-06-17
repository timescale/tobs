package tests

import (
    "fmt"
    "os/exec"
    "syscall"
    "testing"
    "time"
)

func testGrafanaPortForward(t testing.TB, port string) {
    var portforward *exec.Cmd

    if port == "" {
        fmt.Println("Running 'ts-obs grafana port-forward'")
        portforward = exec.Command("/home/archen/go/bin/ts-obs", "grafana", "port-forward")
    } else {
        fmt.Printf("Running 'ts-obs grafana port-forward -p %v'\n", port)
        portforward = exec.Command("/home/archen/go/bin/ts-obs", "grafana", "port-forward", "-p", port)
    }

    err := portforward.Start()
    if err != nil {
        t.Fatal(err)
    }
    time.Sleep(4 * time.Second)
    portforward.Process.Signal(syscall.SIGINT)
}

func testGrafanaGetPass(t testing.TB) {
    var getpass *exec.Cmd

    fmt.Println("Running 'ts-obs grafana get-initial-password'")
    getpass = exec.Command("/home/archen/go/bin/ts-obs", "grafana", "get-initial-password")

    out, err := getpass.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func testGrafanaChangePass(t testing.TB, newpass string) {
    var changepass *exec.Cmd

    fmt.Printf("Running 'ts-obs grafana change-password %v'\n", newpass)
    changepass = exec.Command("/home/archen/go/bin/ts-obs", "grafana", "change-password", newpass)

    out, err := changepass.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func TestGrafana(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Grafana tests")
    }

    testGrafanaPortForward(t, "")
    testGrafanaPortForward(t, "1235")
    testGrafanaPortForward(t, "134")
    testGrafanaPortForward(t, "3")

    testGrafanaGetPass(t)
    testGrafanaChangePass(t, "kraken")
    testGrafanaChangePass(t, "cereal")
    testGrafanaChangePass(t, "23498MSDF(*9389m*(@#M24309mDj")
    testGrafanaGetPass(t)
}
