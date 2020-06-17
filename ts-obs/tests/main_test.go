package tests

import (
    "fmt"
    "os"
    "os/exec"
    "os/signal"
    "testing"
    "time"
)

func startKube() {
    var err error

    fmt.Println("Starting Kubernetes cluster")

    kubestart := exec.Command("minikube", "start")
    err = kubestart.Run()
    if err != nil {
        fmt.Println("Error starting Kubernetes cluster", err)
        os.Exit(1)
    }
}

func installObs() {
    var err error

    fmt.Println("Installing Timescale Observability")

    obsinstall := exec.Command("/home/archen/go/bin/ts-obs", "install")
    err = obsinstall.Run()
    if err != nil {
        fmt.Println("Error installing Timescale Observability:", err)
        os.Exit(1)
    }
}

func stopKube() {
    var err error

    fmt.Println("Stopping Kubernetes cluster")

    kubestop := exec.Command("minikube", "stop")
    err = kubestop.Run()
    if err != nil {
        fmt.Println("Error stopping Kubernetes cluster", err)
        os.Exit(1)
    }
}

func deleteKube() {
    var err error

    fmt.Println("Deleting Kubernetes cluster")
            
    kubedelete := exec.Command("minikube", "delete")
    err = kubedelete.Run()
    if err != nil {
        fmt.Println("Error deleting Kubernetes cluster", err)
        os.Exit(1)
    }
}

func TestMain(m *testing.M) {

    // Signal handling
    sigchan := make(chan os.Signal, 1)
    done := make(chan bool, 1)
    signal.Notify(sigchan, os.Interrupt)
    go func() {
        <- sigchan
        stopKube()
        deleteKube()
        done <- true
        os.Exit(1)
    }()

    startKube()
    installObs()

    fmt.Println("Sleep to wait for pods to initialize")
    time.Sleep(300 * time.Second)

    code := m.Run()

    // Clean up the cluster
    stopKube()
    deleteKube()

    os.Exit(code)
}
