package tests

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"
)

const RELEASE_NAME = "gg"
const NAMESPACE = "ns"

func startKube() {
	var err error

	log.Println("Starting Kubernetes cluster")

	kubestart := exec.Command("minikube", "start")
	err = kubestart.Run()
	if err != nil {
		log.Println("Error starting Kubernetes cluster", err)
		os.Exit(1)
	}
}

func installObs() {
	var err error

	log.Println("Installing Timescale Observability")

	obsinstall := exec.Command("ts-obs", "install", "-n", RELEASE_NAME, "--namespace", NAMESPACE)
	err = obsinstall.Run()
	if err != nil {
		log.Println("Error installing Timescale Observability:", err)
		os.Exit(1)
	}
}

func stopKube() {
	var err error

	log.Println("Stopping Kubernetes cluster")

	kubestop := exec.Command("minikube", "stop")
	err = kubestop.Run()
	if err != nil {
		log.Println("Error stopping Kubernetes cluster", err)
		os.Exit(1)
	}
}

func deleteKube() {
	var err error

	log.Println("Deleting Kubernetes cluster")

	kubedelete := exec.Command("minikube", "delete")
	err = kubedelete.Run()
	if err != nil {
		log.Println("Error deleting Kubernetes cluster", err)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {

	// Signal handling
	sigchan := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		stopKube()
		deleteKube()
		done <- true
		os.Exit(1)
	}()

	startKube()
	installObs()

	time.Sleep(30 * time.Second)

	code := m.Run()

	// Clean up the cluster
	stopKube()
	deleteKube()

	os.Exit(code)
}
