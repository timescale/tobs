package tobs_cli_tests

import (
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
	"testing"
)

func TestPromscale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Promscale tests")
	}

	releaseInfo := test_utils.ReleaseInfo{
		Release:   RELEASE_NAME,
		Namespace: NAMESPACE,
	}

	releaseInfo.TestPromscalePortForward(t, "")
	releaseInfo.TestPromscalePortForward(t, "3421")
}
