package ha_tests

import (
	"log"
	"os"
	"os/signal"
	"testing"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
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
	log.Println("Starting the HA tests....")
	code := m.Run()
	err := test_utils.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	// uninstall tobs
	spec := test_utils.TestUnInstallSpec{
		ReleaseName: RELEASE_NAME,
		Namespace:   NAMESPACE,
		DeleteData:  true,
	}
	spec.TestUninstall(&testing.T{})

	os.Exit(code)
}
