package helm_tests

import (
	"os"
	"os/signal"
	"testing"
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

	// created new helm client
	testNewHelmClient()

	// add or upgrade timescale/tobs helm repo
	testHelmClientAddOrUpdateChartRepoPublic()

	// get tobs helm chart metadata
	testGetChartMetadata()

	// installs the tobs helm chart
	testHelmClientInstallOrUpgradeChart()

	// runs all helm tests
	code := m.Run()

	// post all tests execution verify does the
	// tobs helm installation still exists
	testTobsReleasePostUninstall()

	os.Exit(code)
}
