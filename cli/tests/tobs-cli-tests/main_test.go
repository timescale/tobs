package tobs_cli_tests

import (
	"fmt"
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
var PATH_TO_TEST_VALUES = "./../testdata/main-values.yaml"

func installObs() {
	var err error

	log.Println("Installing The Observability Stack")

	obsinstall := exec.Command(PATH_TO_TOBS, "install", "-n", RELEASE_NAME, "--namespace", NAMESPACE, "--chart-reference", PATH_TO_CHART, "-f", PATH_TO_TEST_VALUES, "--enable-prometheus-ha")
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

	out := exec.Command("helm", "dep", "up", PATH_TO_CHART)
	_, err := out.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	// tests on backEnabled tobs
	// runs it prior to other tests as
	// the tobs installation itself is different
	testBackUpEnabledInstallation(&testing.T{})

	installObs()

	time.Sleep(3 * time.Minute)

	fmt.Println("starting e2e tests post tobs deployment....")
	code := m.Run()

	err = test_utils.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	uninstallsObs()

	// wait for the uninstall to succeed
	// this takes 3 mins because in HA mode
	// we have three 3 Prometheus instances to gracefully shutdown
	// and to avoid flakiness.
	time.Sleep(3 * time.Minute)

	err = test_utils.CheckPVCSExist(RELEASE_NAME, NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

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
