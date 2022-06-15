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
var PATH_TO_TEST_VALUES = "./../testdata/e2e-values.yaml"

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

	// TODO(paulfantom): Consider moving this out of test suite
	log.Println("Updating helm charts...")
	out := exec.Command("helm", "dep", "up", PATH_TO_CHART)
	_, err := out.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("successfully performed backup install tests...")

	installObs()

	// FIXME(paulfantom): This should be converted into polling for instances
	time.Sleep(5 * time.Minute)

	log.Println("starting e2e tests post tobs deployment....")
	code := m.Run()

	err = test_utils.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	uninstallsObs()

	// wait for the uninstall to succeed
	// this takes 5 mins because in HA mode
	// we have 2 Prometheus instances to gracefully shutdown
	// and to avoid flakiness.
	// FIXME(paulfantom): This should be converted into polling for instances
	time.Sleep(5 * time.Minute)

	// TODO(paulfantom): This should be a part of TestUninstall()
	err = test_utils.CheckPVCSExist(RELEASE_NAME, NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func installObs() {
	var err error

	log.Println("Installing The Observability Stack")

	test_utils.ShowAllPods(&testing.T{})

	obsinstall := exec.Command(PATH_TO_TOBS, "install", "--name", RELEASE_NAME, "--namespace", NAMESPACE, "--chart-reference", PATH_TO_CHART, "-f", PATH_TO_TEST_VALUES, "--enable-prometheus-ha")
	out, err := obsinstall.CombinedOutput()
	log.Println(string(out))
	if err != nil {
		log.Println("Error installing The Observability Stack:", err)
		os.Exit(1)
	}
}

func uninstallsObs() {
	log.Println("Uninstalling The Observability Stack")
	obsinstall := exec.Command(PATH_TO_TOBS, "uninstall", "--name", RELEASE_NAME, "--namespace", NAMESPACE, "--delete-data")
	out, err := obsinstall.CombinedOutput()
	log.Println(string(out))
	if err != nil {
		log.Println("Error installing The Observability Stack:", err)
		os.Exit(1)
	}
}
