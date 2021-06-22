package external_db_tests

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

func TestPromscale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Promscale tests")
	}
	fmt.Println("Running Promscale tests for external db setup...")

	// Tests based on pod status & restarts
	promscalePod, err := kubeClient.K8s.KubeGetPods(NAMESPACE, map[string]string{"app": "tobs-promscale"})
	if err != nil {
		log.Println("failed to get promscale pod")
		os.Exit(1)
	}

	if len(promscalePod) > 0 {
		c := 0
		restarts := promscalePod[0].Status.ContainerStatuses[0].RestartCount
		for {
			if promscalePod[0].Status.Phase != "Running" {
				log.Println("failed to validate promscale with external db as it isn't in running state.")
				os.Exit(1)
			}

			r := promscalePod[0].Status.ContainerStatuses[0].RestartCount
			if r != restarts {
				log.Println("failed to validate promscale with external db as it restarts multiple times.")
				os.Exit(1)
			}

			if c == 3 {
				break
			}

			time.Sleep(10 * time.Second)
			c++
		}

	}

	// explicit check Promscale status as Promscale depends on external db
	// credentials configured
	fmt.Println("Promscale is running successfully with external db setup")

	releaseInfo := test_utils.ReleaseInfo{
		Release:   RELEASE_NAME,
		Namespace: NAMESPACE,
	}

	// Tests based on port-forward
	releaseInfo.TestPromscalePortForward(t, "")
	releaseInfo.TestPromscalePortForward(t, "3421")
}
