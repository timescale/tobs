package upgrade_tests

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"

	"github.com/timescale/tobs/cli/pkg/helm"
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

var (
	PATH_TO_TOBS        = "./../../bin/tobs"
	PATH_TO_CHART       = "./../../../chart/"
	PATH_TO_TEST_VALUES = "./../testdata/e2e-values.yaml"
	NAMESPACE           = "ns"
	RELEASE_NAME        = "gg"
	helmClient          helm.Client
)

// TODO(paulfantom): Figure out how to use only previous tobs version to upgrade from.
// As in `master` should be upgradable from previous release
// release X.Y.Z should be upgradable from release X.(Y-1)
// Maybe VERSION file could help in this
const upgradeFromVersion = "0.7.0"

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

	// upgrade takes some time for pods to get into running state
	time.Sleep(2 * time.Minute)

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

func installTobsRecentRelease() {
	// Deploying tobs 0.2.2
	log.Printf("Installing Tobs %s\n", upgradeFromVersion)
	runTsdb := exec.Command(PATH_TO_TOBS, "install", "--name", RELEASE_NAME, "--namespace", NAMESPACE, "--version", upgradeFromVersion)
	_, err := runTsdb.CombinedOutput()
	if err != nil {
		log.Fatalf("Error installing tobs version %s %v:", upgradeFromVersion, err)
	}
}

func upgradeTobsLatest() {
	// Note: The below tobs cmd only deploys TimescaleDB as the test values.yaml is configured
	// to deploy only timescaleDB
	log.Println("Upgrade to Tobs latest")
	runTsdb := exec.Command(PATH_TO_TOBS, "upgrade", "-c", PATH_TO_CHART, "-f", PATH_TO_TEST_VALUES, "--name", RELEASE_NAME, "--namespace", NAMESPACE, "-y")
	output, err := runTsdb.CombinedOutput()
	if err != nil {
		log.Fatalf("Error upgrading tobs to latest version: %s %v", output, err)
	}
}
