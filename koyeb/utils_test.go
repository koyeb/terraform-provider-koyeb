package koyeb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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
		client.ServicePoolsApi.GetServicePool(context.Background(), "pool-uuid").Execute,
		"ServicePool", []string{"DELETING"}, 1, false,
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
		client.ServicePoolsApi.GetServicePool(context.Background(), "pool-uuid").Execute,
		"ServicePool", []string{"DELETING"}, 1, false,
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
		client.ServicePoolsApi.GetServicePool(context.Background(), "pool-uuid").Execute,
		"ServicePool", []string{"DELETING"}, 0, false,
	)
	if err == nil || !strings.Contains(err.Error(), "resource failed to reach target status") {
		t.Fatalf("expected a timeout error for a stuck pool, got %v", err)
	}
}
