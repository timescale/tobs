package ha_tests

import (
	"github.com/timescale/tobs/cli/pkg/k8s"
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
	"log"
	"os"
	"os/signal"
	"testing"
)

var (
	RELEASE_NAME        = "tobs"
	NAMESPACE           = "default"
	PATH_TO_TOBS        = "./../../bin/tobs"
	PATH_TO_CHART       = "./../../../chart/"
	PATH_TO_TEST_VALUES = "./../testdata/main-values.yaml"
	kubeClient          = &test_utils.TestClient{}
)

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
	log.Println("Starting the HA tests....")
	code := m.Run()
	err := kubeClient.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}
