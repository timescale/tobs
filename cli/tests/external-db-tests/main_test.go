package external_db_tests

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

var RELEASE_NAME = "tobs"
var NAMESPACE = "default"
var PATH_TO_TOBS = "./../../bin/tobs"
var PATH_TO_CHART = "./../../../chart/"

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

	installTimescaleDB()

	// As the timescaleDB setup takes 1 min to start up,
	// sleep for 1 min
	time.Sleep(60 * time.Second)

	runTobsWithoutTSDB()

	// As the tobs setup takes 30 secs to start up,
	// sleep for 30 secs
	time.Sleep(30 * time.Second)

	fmt.Println("Successfully deployed the tobs with external db setup...")

	code := m.Run()

	err := test_utils.CheckPodsRunning(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(code)
}

func installTimescaleDB() {
	// Note: The below tobs cmd only deploys TimescaleDB as the test values.yaml is configured
	// to deploy only timescaleDB
	log.Println("Installing TimescaleDB")
	runTsdb := exec.Command(PATH_TO_TOBS, "install", "-c", PATH_TO_CHART, "-f", "./../testdata/f5.yaml", "--name", "external-db-tests")
	err := runTsdb.Run()
	if err != nil {
		log.Println("Error installing timescaleDB:", err)
		os.Exit(1)
	}
}

func runTobsWithoutTSDB() {
	var err error
	log.Println("Installing The Observability Stack")

	ip, err := test_utils.CreateTimescaleDBNodePortService(NAMESPACE)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created nodeport service for timescaleDB, connecting using ip %s\n", ip)

	obsinstall := exec.Command(PATH_TO_TOBS, "install", "--external-timescaledb-uri",
		"postgres://postgres:tea@"+ip+"/postgres?sslmode=require", "-c", PATH_TO_CHART)
	err = obsinstall.Run()
	if err != nil {
		log.Println("Error installing The Observability Stack without TimescaleDB:", err)
		os.Exit(1)
	}
}
