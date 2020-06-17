package tests

import (
    "fmt"
    "os"
    "os/exec"
    "syscall"
    "testing"
    "time"
)

func testTimescaleGetPassword(t testing.TB, user string) {
    var getpass *exec.Cmd

    if user == "" {
        fmt.Println("Running 'ts-obs timescaledb get-password'")
        getpass = exec.Command("/home/archen/go/bin/ts-obs", "timescaledb", "get-password")
    } else {
        fmt.Printf("Running 'ts-obs timescaledb get-password -u %v'\n", user)
        getpass = exec.Command("/home/archen/go/bin/ts-obs", "timescaledb", "get-password", "-u", user)
    }

    out, err := getpass.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func testTimescalePortForward(t testing.TB, port string) {
    var portforward *exec.Cmd

    if port == "" {
        fmt.Println("Running 'ts-obs timescaledb port-forward'")
        portforward = exec.Command("/home/archen/go/bin/ts-obs", "timescaledb", "port-forward")
    } else {
        fmt.Printf("Running 'ts-obs timescaledb port-forward -p %v'\n", port)
        portforward = exec.Command("/home/archen/go/bin/ts-obs", "timescaledb", "port-forward", "-p", port)
    }

    err := portforward.Start()
    if err != nil {
        t.Fatal(err)
    }
    time.Sleep(4 * time.Second)
    portforward.Process.Signal(syscall.SIGINT)

}

func testTimescaleConnect(t testing.TB, master bool, user, password string) {
    var connect *exec.Cmd

    if master {
        fmt.Println("Running 'ts-obs timescaledb connect -m'")
        connect = exec.Command("/home/archen/go/bin/ts-obs", "timescaledb", "connect", "-m")
    } else {
        if user == "" {
            fmt.Printf("Running 'ts-obs timescaledb connect -p %v'\n", password)
            connect = exec.Command("/home/archen/go/bin/ts-obs", "timescaledb", "connect", "-p", password)
        } else {
            fmt.Printf("Running 'ts-obs timescaledb connect -u %v -p %v'\n", user, password)
            connect = exec.Command("/home/archen/go/bin/ts-obs", "timescaledb", "connect", "-u", user, "-p", password)
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
    testTimescalePortForward(t, "10")
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
