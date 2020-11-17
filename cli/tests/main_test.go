package tests

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"
)

var RELEASE_NAME = "gg"
var NAMESPACE = "ns"

func installObs() {
	var err error

	log.Println("Installing The Observability Stack")

	obsinstall := exec.Command("./../bin/tobs", "install", "-n", RELEASE_NAME, "--namespace", NAMESPACE, "--chart-reference", "../../chart")
	err = obsinstall.Run()
	if err != nil {
		log.Println("Error installing The Observability Stack:", err)
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
		done <- true
		os.Exit(1)
	}()

	installObs()

	time.Sleep(30 * time.Second)

	code := m.Run()

	os.Exit(code)
}
