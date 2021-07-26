package tobs_cli_tests

import (
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
	"testing"
)

func TestTimescale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TimescaleDB tests")
	}

	releaseInfo := test_utils.ReleaseInfo{
		Release:   RELEASE_NAME,
		Namespace: NAMESPACE,
	}

	releaseInfo.TestTimescalePortForward(t, "")
	releaseInfo.TestTimescalePortForward(t, "5432")
	releaseInfo.TestTimescalePortForward(t, "1789")

	releaseInfo.TestTimescaleConnect(t, true, "")
	releaseInfo.TestTimescaleConnect(t, false, "")
	releaseInfo.TestTimescaleConnect(t, false, "postgres")
	releaseInfo.TestTimescaleConnect(t, false, "postgres")
	releaseInfo.TestTimescaleConnect(t, false, "admin")
	releaseInfo.TestTimescaleConnect(t, false, "admin")
}
