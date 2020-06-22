package tests

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func testRetentionSetDefault(t testing.TB, period int) {
	var set *exec.Cmd

	t.Logf("Running 'ts-obs metrics retention set-default %d'\n", period)
	set = exec.Command("ts-obs", "metrics", "retention", "set-default", strconv.Itoa(period))

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testRetentionSet(t testing.TB, metric string, period int) {
	var set *exec.Cmd

	t.Logf("Running 'ts-obs metrics retention set %v %d'\n", metric, period)
	set = exec.Command("ts-obs", "metrics", "retention", "set", metric, strconv.Itoa(period))

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testRetentionReset(t testing.TB, metric string) {
	var reset *exec.Cmd

	t.Logf("Running 'ts-obs metrics retention reset %v'\n", metric)
	reset = exec.Command("ts-obs", "metrics", "retention", "reset", metric)

	out, err := reset.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testChunkIntervalSetDefault(t testing.TB, interval string) {
	var set *exec.Cmd

	t.Logf("Running 'ts-obs metrics chunk-interval set-default %v'\n", interval)
	set = exec.Command("ts-obs", "metrics", "chunk-interval", "set-default", interval)

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testChunkIntervalSet(t testing.TB, metric, interval string) {
	var set *exec.Cmd

	t.Logf("Running 'ts-obs metrics chunk-interval set %v %v'\n", metric, interval)
	set = exec.Command("ts-obs", "metrics", "chunk-interval", "set", metric, interval)

	out, err := set.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func testChunkIntervalReset(t testing.TB, metric string) {
	var reset *exec.Cmd

	t.Logf("Running 'ts-obs metrics chunk-interval reset %v'\n", metric)
	reset = exec.Command("ts-obs", "metrics", "chunk-interval", "reset", metric)

	out, err := reset.CombinedOutput()
	if err != nil {
		t.Logf(string(out))
		t.Fatal(err)
	}
}

func verifyRetentionPeriod(t testing.TB, metric string, expectedDuration time.Duration) {
	var durS int
	var dur time.Duration

	portforward := exec.Command("kubectl", "--namespace", "default", "port-forward", "ts-obs-timescaledb-0", "5433:5432")
	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	os.Setenv("PGPASSWORD_POSTGRES", "tea")

	pool, err := pgxpool.Connect(context.Background(), "postgres://postgres:"+os.Getenv("PGPASSWORD_POSTGRES")+"@localhost:5433/postgres")
	if err != nil {
		t.Fatal(err)
	}

	err = pool.QueryRow(context.Background(),
		`SELECT EXTRACT(epoch FROM _prom_catalog.get_metric_retention_period($1))`,
		metric).Scan(&durS)
	if err != nil {
		t.Error(err)
	}
	dur = time.Duration(durS) * time.Second

	if dur != expectedDuration {
		t.Fatalf("Unexpected retention period for table %v: got %v want %v", metric, dur, expectedDuration)
	}
	pool.Close()
}

func verifyChunkInterval(t testing.TB, tableName string, expectedDuration time.Duration) {
	var intervalLength int64

	portforward := exec.Command("kubectl", "--namespace", "default", "port-forward", "ts-obs-timescaledb-0", "5433:5432")
	err := portforward.Start()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	os.Setenv("PGPASSWORD_POSTGRES", "tea")

	pool, err := pgxpool.Connect(context.Background(), "postgres://postgres:"+os.Getenv("PGPASSWORD_POSTGRES")+"@localhost:5433/postgres")
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
		t.Error(err)
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

	testRetentionSetDefault(t, 10)
	verifyRetentionPeriod(t, "container_last_seen", 10*24*time.Hour)

	testRetentionReset(t, "up")
	verifyRetentionPeriod(t, "up", 10*24*time.Hour)

	testRetentionSet(t, "node_load15", 9)
	verifyRetentionPeriod(t, "node_load15", 9*24*time.Hour)

	testRetentionSet(t, "up", 2)
	verifyRetentionPeriod(t, "up", 2*24*time.Hour)

	testRetentionSet(t, "kube_pod_status_phase", 32)
	verifyRetentionPeriod(t, "kube_pod_status_phase", 32*24*time.Hour)

	testRetentionReset(t, "up")
	verifyRetentionPeriod(t, "up", 10*24*time.Hour)

	testRetentionReset(t, "node_load5")
	verifyRetentionPeriod(t, "node_load15", 10*24*time.Hour)

	testRetentionSetDefault(t, 11)
	verifyRetentionPeriod(t, "go_info", 11*24*time.Hour)

	testChunkIntervalSet(t, "container_last_seen", "23m45s")
	verifyChunkInterval(t, "container_last_seen", (23*60+45)*time.Second)

	testChunkIntervalSetDefault(t, "62m3s")
	verifyChunkInterval(t, "node_load15", (62*60+3)*time.Second)

	testChunkIntervalSet(t, "go_info", "3403s")
	verifyChunkInterval(t, "go_info", 3403*time.Second)

	testChunkIntervalReset(t, "go_info")
	verifyChunkInterval(t, "go_info", (62*60+3)*time.Second)

	testChunkIntervalSet(t, "kube_job_info", "8h24m")
	verifyChunkInterval(t, "kube_job_info", (8*60+24)*time.Minute)

	testChunkIntervalSetDefault(t, "23h")
	verifyChunkInterval(t, "kube_pod_status_phase", (23)*time.Hour)

	testChunkIntervalReset(t, "go_threads")
	verifyChunkInterval(t, "go_threads", (23)*time.Hour)

}
