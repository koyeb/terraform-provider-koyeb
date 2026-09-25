package koyeb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestWaitForResourceStatusVolumeSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"volume":{"id":"vol-uuid","status":"PERSISTENT_VOLUME_STATUS_DELETED"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.PersistentVolumesApi.GetPersistentVolume(context.Background(), "vol-uuid").Execute,
		"Volume", []string{"PERSISTENT_VOLUME_STATUS_DELETED", "PERSISTENT_VOLUME_STATUS_DELETING"}, time.Minute, false,
	)
	if err != nil {
		t.Fatalf("expected the DELETED volume to satisfy the wait, got %s", err)
	}
}

func TestWaitForResourceStatusVolumeTimesOutWhenStuck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"volume":{"id":"vol-uuid","status":"PERSISTENT_VOLUME_STATUS_READY"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.PersistentVolumesApi.GetPersistentVolume(context.Background(), "vol-uuid").Execute,
		"Volume", []string{"PERSISTENT_VOLUME_STATUS_DELETED", "PERSISTENT_VOLUME_STATUS_DELETING"}, 0, false,
	)
	if err == nil || !strings.Contains(err.Error(), "resource failed to reach target status") {
		t.Fatalf("expected a timeout error for a stuck volume, got %v", err)
	}
}

func TestWaitForResourceStatusServicePoolSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","status":"DELETING"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.ServicePoolsApi.GetServicePool(context.Background(), "pool-uuid").Execute,
		"ServicePool", []string{"DELETING"}, time.Minute, false,
	)
	if err != nil {
		t.Fatalf("expected the DELETING pool to satisfy the wait, got %s", err)
	}
}

func TestWaitForResourceStatusServicePoolTreats404AsGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.ServicePoolsApi.GetServicePool(context.Background(), "pool-uuid").Execute,
		"ServicePool", []string{"DELETING"}, time.Minute, false,
	)
	if err != nil {
		t.Fatalf("expected a 404 to count as destroyed, got %s", err)
	}
}

func TestWaitForResourceStatusServicePoolTimesOutWhenStuck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","status":"READY"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	// A zero timeout exhausts the window immediately: a pool that never
	// reaches the target must surface the timeout error.
	err := waitForResourceStatus(
		context.Background(),
		client.ServicePoolsApi.GetServicePool(context.Background(), "pool-uuid").Execute,
		"ServicePool", []string{"DELETING"}, 0, false,
	)
	if err == nil || !strings.Contains(err.Error(), "resource failed to reach target status") {
		t.Fatalf("expected a timeout error for a stuck pool, got %v", err)
	}
}

func TestWaitForResourceStatusSnapshotSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"snapshot":{"id":"snap-uuid","status":"SNAPSHOT_STATUS_DELETED"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.SnapshotsApi.GetSnapshot(context.Background(), "snap-uuid").Execute,
		"Snapshot", []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}, time.Minute, false,
	)
	if err != nil {
		t.Fatalf("expected the DELETED snapshot to satisfy the wait, got %s", err)
	}
}

func TestWaitForResourceStatusSnapshotTimesOutWhenStuck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"snapshot":{"id":"snap-uuid","status":"SNAPSHOT_STATUS_CREATING"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.SnapshotsApi.GetSnapshot(context.Background(), "snap-uuid").Execute,
		"Snapshot", []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}, 0, false,
	)
	if err == nil || !strings.Contains(err.Error(), "resource failed to reach target status") {
		t.Fatalf("expected a timeout error for a stuck snapshot, got %v", err)
	}
}

func TestWaitForResourceStatusSnapshotTreats404AsGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.SnapshotsApi.GetSnapshot(context.Background(), "snap-uuid").Execute,
		"Snapshot", []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}, time.Minute, false,
	)
	if err != nil {
		t.Fatalf("expected a 404 to count as destroyed, got %s", err)
	}
}

func TestWaitForResourceStatusHonorsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","status":"PROVISIONING"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := waitForResourceStatus(
		ctx,
		client.ServicePoolsApi.GetServicePool(context.Background(), "pool-uuid").Execute,
		"ServicePool", []string{"READY"}, time.Minute, false,
	)
	if err == nil || !strings.Contains(err.Error(), "cancel") {
		t.Fatalf("expected a cancellation error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("expected the wait to unblock immediately, took %s", elapsed)
	}
}

func TestWaitForResourceStatusAppSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"app":{"id":"app-uuid","status":"DELETED"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForResourceStatus(
		context.Background(),
		client.AppsApi.GetApp(context.Background(), "app-uuid").Execute,
		"App", []string{"DELETED", "DELETING"}, time.Minute, false,
	)
	if err != nil {
		t.Fatalf("expected the DELETED app to satisfy the wait, got %s", err)
	}
}
