package installation_tests

import (
	"log"
	"os"
	"os/signal"
	"testing"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

var RELEASE_NAME = "gg"
var NAMESPACE = "ns"
var PATH_TO_TOBS = "./../../bin/tobs"
var PATH_TO_CHART = "./../../../chart/"
var PATH_TO_TEST_VALUES = "./../testdata/main-values.yaml"

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

	code := m.Run()

	err := test_utils.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}
