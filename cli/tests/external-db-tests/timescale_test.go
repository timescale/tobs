package external_db_tests

import (
	"fmt"
	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
	"testing"
)

func TestTimescale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TimescaleDB tests")
	}

	fmt.Println("Performing TimescaleDB tests for external db setup...")
	releaseInfo := test_utils.ReleaseInfo{
		Release:   RELEASE_NAME,
		Namespace: NAMESPACE,
	}

	//releaseInfo.TestTimescaleGetPassword(t, "")
	//releaseInfo.TestTimescaleChangePassword(t, "", "postgres", "battery")
	//releaseInfo.VerifyTimescalePassword(t, "postgres", "battery")
	//releaseInfo.TestTimescaleGetPassword(t, "admin")
	//releaseInfo.TestTimescaleChangePassword(t, "admin", "", "chips")
	//releaseInfo.VerifyTimescalePassword(t, "admin", "chips")

	releaseInfo.TestTimescaleConnect(t, true, "")
	releaseInfo.TestTimescaleConnect(t, false, "")
	releaseInfo.TestTimescaleConnect(t, false, "postgres")
	releaseInfo.TestTimescaleConnect(t, false, "postgres")
	releaseInfo.TestTimescaleConnect(t, false, "admin")
	releaseInfo.TestTimescaleConnect(t, false, "admin")
}
