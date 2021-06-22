package tobs_cli_tests

import (
	"fmt"
	"github.com/timescale/tobs/cli/pkg/k8s"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

var (
	RELEASE_NAME        = "gg"
	NAMESPACE           = "ns"
	PATH_TO_TOBS        = "./../../bin/tobs"
	PATH_TO_CHART       = "./../../../chart/"
	PATH_TO_TEST_VALUES = "./../testdata/main-values.yaml"
	kubeClient          = &test_utils.TestClient{}
)

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

	kubeClient.K8s, _ = k8s.NewClient()
	// tests on backEnabled tobs
	// runs it prior to other tests as
	// the tobs installation itself is different
	testBackUpEnabledInstallation(&testing.T{})

	installObs()

	time.Sleep(3 * time.Minute)

	fmt.Println("starting e2e tests post tobs deployment....")
	code := m.Run()

	err := kubeClient.CheckPodsRunning(NAMESPACE)
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
