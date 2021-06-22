package ha_tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

func TestHASetup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping upgrade tests")
	}

	kubeClient.K8s, _ = k8s.NewClient()
	installTobs(t)
	fmt.Println("Successfully installed ha test setup.....")
	time.Sleep(4 * time.Minute)

	err := kubeClient.CheckPodsRunning(NAMESPACE)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("All pods are in running state....")
	test_utils.PortForwardPromscale(t, RELEASE_NAME, NAMESPACE)
	var oldLeader, newLeader string
	// Prometheus leader on startup
	oldLeader = findCurrentLeader(t)
	fmt.Println("Old Leader: ", oldLeader)
	retries := 0
	for {
		i := 0
		for i < 5 {
			_ = kubeClient.DeletePod(oldLeader, NAMESPACE)
			i++
			time.Sleep(3 * time.Second)
		}
		// Prometheus new leader after
		// shutting down the old Prometheus leader
		newLeader = findCurrentLeader(t)
		fmt.Println("New Leader: ", newLeader)
		if oldLeader != newLeader {
			fmt.Printf("Leader succesffuly swicthed from Prometheus instance %s to Prometheus instance %s\n", oldLeader, newLeader)
			break
		}

		// Every retry consumes 3 secs sleep * 5 attempts = 15 secs
		// In Promscale Prometheus leader change over happens when
		// the last write request is older than 30s. Approx after 3 retries i.e. 45 secs
		// the change-over should happen :)
		if retries == 10 {
			t.Fatal("Leader switch over doesn't happen after multiple retries....")
		}
		retries++
	}

	time.Sleep(30 * time.Second)
	// re-verify the new Prometheus leader
	if newLeader == findCurrentLeader(t) {
		fmt.Println("Re-verified the leader and it's the same new leader: ", newLeader)
	} else {
		t.Fatalf("failed to re-verify Promscale HA switch over....")
	}
}

func installTobs(t *testing.T) {
	t.Log("Installing Tobs...")
	runTsdb := exec.Command(PATH_TO_TOBS, "install", "--name", RELEASE_NAME, "--namespace", NAMESPACE, "-c", PATH_TO_CHART, "-f", PATH_TO_TEST_VALUES, "--enable-prometheus-ha")
	_, err := runTsdb.CombinedOutput()
	if err != nil {
		t.Fatalf("Error installing tobs version %v:", err)
	}
}

func getPromscaleMetrics(t *testing.T) string {
	res, err := http.Get("http://localhost:9201/metrics")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	return string(data)
}

func findHAReplicas(t *testing.T) map[string]string {
	data := getPromscaleMetrics(t)
	haReplicas := make(map[string]string)
	samples := strings.Split(data, "\n")
	for _, s := range samples {
		// check is HA metric
		isHAMetric := strings.Contains(s, "promscale_ha_cluster_leader_info{")
		if isHAMetric {
			d := strings.Split(s, " ")
			// extract the Prometheus replica name from HA metric labelSet
			replica := strings.Split(d[0], "replica=\"")
			replicaName := strings.TrimSuffix(replica[1], "\"}")
			haReplicas[replicaName] = d[1]
		}
	}

	return haReplicas
}

func findCurrentLeader(t *testing.T) string {
	haInfo := findHAReplicas(t)
	for k, v := range haInfo {
		if v == "1" {
			return k
		}
	}
	return ""
}
