package tobs_cli_tests

import (
	"testing"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

func TestTimescaleSuper(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TimescaleDB Super User tests")
	}

	releaseInfo := test_utils.ReleaseInfo{
		Release:   RELEASE_NAME,
		Namespace: NAMESPACE,
	}

	releaseInfo.TestTimescaleGetPassword(t)
	releaseInfo.TestTimescaleChangePassword(t, "battery")
	releaseInfo.VerifyTimescalePassword(t, "battery")
	releaseInfo.TestTimescaleGetPassword(t)
	releaseInfo.TestTimescaleChangePassword(t, "chips")
	releaseInfo.VerifyTimescalePassword(t, "chips")

	releaseInfo.TestTimescaleSuperUserConnect(t, true)
	releaseInfo.TestTimescaleSuperUserConnect(t, false)
	releaseInfo.TestTimescaleSuperUserConnect(t, false)
	releaseInfo.TestTimescaleSuperUserConnect(t, false)
	releaseInfo.TestTimescaleSuperUserConnect(t, false)
	releaseInfo.TestTimescaleSuperUserConnect(t, false)
}
