package tobs_cli_tests

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

var RELEASE_NAME = "gg"
var NAMESPACE = "ns"
var PATH_TO_TOBS = "./../../bin/tobs"
var PATH_TO_CHART = "./../../../chart/"

func installObs() {
	var err error

	log.Println("Installing The Observability Stack")

	obsinstall := exec.Command(PATH_TO_TOBS, "install", "-n", RELEASE_NAME, "--namespace", NAMESPACE, "--chart-reference", PATH_TO_CHART)
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

	err := test_utils.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	uninstallsObs()

	os.Exit(code)
}

func uninstallsObs() {
	log.Println("Uninstalling The Observability Stack")
	obsinstall := exec.Command(PATH_TO_TOBS, "uninstall", "-n", RELEASE_NAME, "--namespace", NAMESPACE, "--delete-data")
	err := obsinstall.Run()
	if err != nil {
		log.Println("Error installing The Observability Stack:", err)
		os.Exit(1)
	}
}
