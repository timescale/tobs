package tobs_cli_tests

import (
	"os"
	"testing"

	test_utils "github.com/timescale/tobs/cli/tests/test-utils"
)

type backupDetails struct {
	key   string
	value string
}

func testBackUpEnabledInstallation(t *testing.T) {
	t.Log("performing backup enabled tests....")
	releaseName := "testbackup"
	namespace := "testbackup"
	bucket := backupDetails{
		key:   "PGBACKREST_REPO1_S3_BUCKET",
		value: "abc",
	}

	endPoint := backupDetails{
		key:   "PGBACKREST_REPO1_S3_ENDPOINT",
		value: "def",
	}

	region := backupDetails{
		key:   "PGBACKREST_REPO1_S3_REGION",
		value: "ghi",
	}

	key := backupDetails{
		key:   "PGBACKREST_REPO1_S3_KEY",
		value: "jkl",
	}

	secret := backupDetails{
		key:   "PGBACKREST_REPO1_S3_KEY_SECRET",
		value: "mno",
	}

	os.Setenv(bucket.key, bucket.value)
	os.Setenv(endPoint.key, endPoint.value)
	os.Setenv(region.key, region.value)
	os.Setenv(key.key, key.value)
	os.Setenv(secret.key, secret.value)
	i := test_utils.TestInstallSpec{
		PathToChart:  PATH_TO_CHART,
		ReleaseName:  releaseName,
		Namespace:    namespace,
		PathToValues: PATH_TO_TEST_VALUES,
		EnableBackUp: true,
		SkipWait:     true,
		OnlySecrets:  false,
	}
	i.TestInstall(t)
	sec, err := kubeClient.GetTSDBBackUpSecret(releaseName, namespace)
	if err != nil {
		t.Logf("Error while finding timescaleDB backup secret. After installting tobs with backup enabled.")
		t.Fatal(err)
	}

	if string(sec.Data[bucket.key]) != bucket.value || string(sec.Data[endPoint.key]) != endPoint.value ||
		string(sec.Data[region.key]) != region.value || string(sec.Data[key.key]) != key.value ||
		string(sec.Data[secret.key]) != secret.value {
		t.Fatal("Error while evaluating secret data in pgbackrest secret the data provided in envs is not matching with data in secret.")
	}

	u := test_utils.TestUnInstallSpec{
		ReleaseName: releaseName,
		Namespace:   namespace,
		DeleteData:  true,
	}
	u.TestUninstall(t)

	_, err = kubeClient.GetTSDBBackUpSecret(releaseName, namespace)
	// here we expect an error after uninstalling the pgbackrest secret shouldn't be found
	if err == nil {
		t.Fatal("Uninstalling backup enabled tobs deployment failed to delete pgbackrest secret")
	}
}
