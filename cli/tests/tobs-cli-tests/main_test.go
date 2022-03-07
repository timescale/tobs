package tobs_cli_tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
)

var RELEASE_NAME = "gg"
var NAMESPACE = "ns"
var PATH_TO_TOBS = "./../../bin/tobs"
var PATH_TO_CHART = "./../../../chart/"
var PATH_TO_TEST_VALUES = "./../testdata/e2e-values.yaml"

func installObs() {
	var err error

	log.Println("Installing The Observability Stack")

	test_utils.ShowAllPods(&testing.T{})

	obsinstall := exec.Command(PATH_TO_TOBS, "install", "--name", RELEASE_NAME, "--namespace", NAMESPACE, "--chart-reference", PATH_TO_CHART, "-f", PATH_TO_TEST_VALUES, "--enable-prometheus-ha")
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
	// tests on backupEnabled tobs
	// runs it prior to other tests as
	// the tobs installation itself is different
	testBackupEnabledInstallation(&testing.T{})

	log.Println("successfully performed backup install tests...")

	installObs()

	err = test_utils.WaitForPodsReady(context.Background(), NAMESPACE, 10*time.Minute, 3,
		metav1.ListOptions{
			LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
				"app.kubernetes.io/name": "promscale-connector",
			})).String(),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("starting e2e tests post tobs deployment....")
	code := m.Run()

	err = test_utils.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	uninstallsObs()

	// Wait untill all pods are removed
	err = test_utils.WaitForPodsReady(context.Background(), NAMESPACE, 15*time.Minute, 0,
		metav1.ListOptions{
			LabelSelector: fields.SelectorFromSet(fields.Set(map[string]string{
				"release": RELEASE_NAME,
			})).String(),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = test_utils.CheckPVCSExist(RELEASE_NAME, NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func uninstallsObs() {
	log.Println("Uninstalling The Observability Stack")
	obsinstall := exec.Command(PATH_TO_TOBS, "uninstall", "--name", RELEASE_NAME, "--namespace", NAMESPACE, "--delete-data")
	err := obsinstall.Run()
	if err != nil {
		log.Println("Error installing The Observability Stack:", err)
		os.Exit(1)
	}
}
