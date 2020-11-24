package tests

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/timescale/tobs/cli/pkg/k8s"

	"github.com/jackc/pgx/v4/pgxpool"
)

var PASS string

func testRetentionSetDefault(t testing.TB, period int, user, dbname string) {
	cmds := []string{"metrics", "retention", "set-default", strconv.Itoa(period), "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	set := exec.Command("./../bin/tobs", cmds...)

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testRetentionSet(t testing.TB, metric string, period int, user, dbname string) {
	cmds := []string{"metrics", "retention", "set", metric, strconv.Itoa(period), "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	set := exec.Command("./../bin/tobs", cmds...)

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testRetentionReset(t testing.TB, metric, user, dbname string) {
	cmds := []string{"metrics", "retention", "reset", metric, "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	reset := exec.Command("./../bin/tobs", cmds...)

	out, err := reset.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testRetentionGet(t testing.TB, metric string, expectedDays int64, user, dbname string) {
	cmds := []string{"metrics", "retention", "get", metric, "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	get := exec.Command("./../bin/tobs", cmds...)

	out, err := get.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	tokens := strings.Fields(string(out))
	days, err := strconv.ParseInt(tokens[len(tokens)-2], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	if days != expectedDays {
		t.Fatalf("Unexpected retention period for table %v: got %v want %v", metric, days, expectedDays)
	}
}

func testChunkIntervalSetDefault(t testing.TB, interval, user, dbname string) {
	cmds := []string{"metrics", "chunk-interval", "set-default", interval, "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	set := exec.Command("./../bin/tobs", cmds...)

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testChunkIntervalSet(t testing.TB, metric, interval, user, dbname string) {
	cmds := []string{"metrics", "chunk-interval", "set", metric, interval, "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	set := exec.Command("./../bin/tobs", cmds...)

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testChunkIntervalReset(t testing.TB, metric, user, dbname string) {
	cmds := []string{"metrics", "chunk-interval", "reset", metric, "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	reset := exec.Command("./../bin/tobs", cmds...)

	out, err := reset.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testChunkIntervalGet(t testing.TB, metric string, expectedDuration time.Duration, user, dbname string) {
	cmds := []string{"metrics", "chunk-interval", "get", metric, "-n", RELEASE_NAME, "--namespace", NAMESPACE}
	if user != "" {
		cmds = append(cmds, "-U", user)
	}
	if dbname != "" {
		cmds = append(cmds, "-d", dbname)
	}

	t.Logf("Running '%v'", "tobs "+strings.Join(cmds, " "))
	get := exec.Command("./../bin/tobs", cmds...)

	out, err := get.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}

	tokens := strings.Fields(string(out))
	duration, err := time.ParseDuration(tokens[len(tokens)-1])
	if err != nil {
		t.Fatal(err)
	}
	if duration.Round(time.Hour) != expectedDuration.Round(time.Hour) {
		t.Errorf("Unexpected chunk interval for table %v: got %v want %v", metric, duration, expectedDuration)
	}
}

func verifyRetentionPeriod(t testing.TB, metric string, expectedDuration time.Duration) {
	var durS int
	var dur time.Duration

	pool, err := pgxpool.Connect(context.Background(), "postgres://postgres:"+PASS+"@localhost:5433/postgres")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.QueryRow(context.Background(),
		`SELECT EXTRACT(epoch FROM _prom_catalog.get_metric_retention_period($1))`,
		metric).Scan(&durS)
	if err != nil {
		t.Fatal(err)
	}
	dur = time.Duration(durS) * time.Second

	if dur != expectedDuration {
		t.Fatalf("Unexpected retention period for table %v: got %v want %v", metric, dur, expectedDuration)
	}
	pool.Close()
}

func verifyChunkInterval(t testing.TB, tableName string, expectedDuration time.Duration) {
	var intervalLength int64

	pool, err := pgxpool.Connect(context.Background(), "postgres://postgres:"+PASS+"@localhost:5433/postgres")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.QueryRow(context.Background(),
		`SELECT d.interval_length
	 FROM _timescaledb_catalog.hypertable h
	 INNER JOIN LATERAL
	 (SELECT dim.interval_length FROM _timescaledb_catalog.dimension dim WHERE dim.hypertable_id = h.id ORDER BY dim.id LIMIT 1) d
	    ON (true)
	 WHERE table_name = $1`,
		tableName).Scan(&intervalLength)
	if err != nil {
		t.Fatal(err)
	}

	dur := time.Duration(time.Duration(intervalLength) * time.Microsecond)
	if dur.Round(time.Hour) != expectedDuration.Round(time.Hour) {
		t.Errorf("Unexpected chunk interval for table %v: got %v want %v", tableName, dur, expectedDuration)
	}
	pool.Close()
}

func TestMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping metrics tests")
	}

	secret, err := k8s.KubeGetSecret(NAMESPACE, RELEASE_NAME+"-timescaledb-passwords")
	if err != nil {
		t.Fatal(err)
	}

	if bytepass, exists := secret.Data["postgres"]; exists {
		PASS = string(bytepass)
	} else {
		t.Fatal(errors.New("user not found"))
	}

	podName, err := k8s.KubeGetPodName(NAMESPACE, map[string]string{"release": RELEASE_NAME, "role": "master"})
	if err != nil {
		t.Fatal(err)
	}

	stdout := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	_, err = k8s.KubePortForwardPod(NAMESPACE, podName, 5433, 5432)
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = stdout

	testRetentionSetDefault(t, 10, "", "")
	verifyRetentionPeriod(t, "container_last_seen", 10*24*time.Hour)

	testRetentionReset(t, "up", "", "")
	verifyRetentionPeriod(t, "up", 10*24*time.Hour)

	testRetentionSet(t, "node_load15", 9, "", "postgres")
	verifyRetentionPeriod(t, "node_load15", 9*24*time.Hour)

	testRetentionSet(t, "up", 2, "postgres", "")
	verifyRetentionPeriod(t, "up", 2*24*time.Hour)

	testRetentionSet(t, "kube_pod_status_phase", 32, "", "postgres")
	verifyRetentionPeriod(t, "kube_pod_status_phase", 32*24*time.Hour)

	testRetentionReset(t, "up", "", "")
	verifyRetentionPeriod(t, "up", 10*24*time.Hour)

	testRetentionReset(t, "node_load5", "", "")
	verifyRetentionPeriod(t, "node_load15", 10*24*time.Hour)

	testRetentionSetDefault(t, 11, "", "postgres")
	verifyRetentionPeriod(t, "go_info", 11*24*time.Hour)

	testRetentionGet(t, "node_load5", 11, "", "")

	testChunkIntervalSet(t, "container_last_seen", "23m45s", "", "")
	verifyChunkInterval(t, "container_last_seen", (23*60+45)*time.Second)

	testChunkIntervalSetDefault(t, "62m3s", "", "")
	verifyChunkInterval(t, "node_load15", (62*60+3)*time.Second)

	testChunkIntervalSet(t, "go_info", "3403s", "postgres", "postgres")
	verifyChunkInterval(t, "go_info", 3403*time.Second)

	testChunkIntervalGet(t, "node_load15", (62*60+3)*time.Second, "", "")

	testChunkIntervalReset(t, "go_info", "", "postgres")
	verifyChunkInterval(t, "go_info", (62*60+3)*time.Second)

	testChunkIntervalSet(t, "kube_job_info", "8h24m", "", "")
	verifyChunkInterval(t, "kube_job_info", (8*60+24)*time.Minute)

	testChunkIntervalSetDefault(t, "23h", "", "postgres")
	verifyChunkInterval(t, "kube_pod_status_phase", (23)*time.Hour)

	testChunkIntervalReset(t, "go_threads", "", "")
	verifyChunkInterval(t, "go_threads", (23)*time.Hour)

}
