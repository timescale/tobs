package external_db_tests

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"testing"
	"time"
)

var RELEASE_NAME = "tobs"
var NAMESPACE = "default"

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

	time.Sleep(60 * time.Second)

	runTobsWithoutTSDB()

	// As the tobs setup takes 30 secs to start up,
	// sleep for 30 secs
	time.Sleep(30 * time.Second)

	code := m.Run()

	os.Exit(code)
}

func installTimescaleDB() {
	// Note: The below tobs cmd only deploy deploys TimescaleDB as the test chart is configured
	// to deploy only timescaleDB
	runTsdb := exec.Command("./../../bin/tobs", "install", "-c", "./../testdata/deploy-timescaledb-chart/", "--name", "external-db-tests")
	err := runTsdb.Run()
	if err != nil {
		fmt.Println(err.Error())
		log.Println("Error installing timescaleDB:", err)
		os.Exit(1)
	}
}

func runTobsWithoutTSDB() {
	var err error
	log.Println("Installing The Observability Stack")

	obsinstall := exec.Command("./../../bin/tobs", "install", "--external-timescaledb-uri",
		"postgres://postgres:tea@external-db-tests.default.svc.cluster.local:5432/postgres?sslmode=require", "-c", "./../../../chart/")
	err = obsinstall.Run()
	if err != nil {
		log.Println("Error installing The Observability Stack without TimescaleDB:", err)
		os.Exit(1)
	}
}
