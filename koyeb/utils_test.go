package koyeb

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestWaitForStatusVolumeSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"volume":{"id":"vol-uuid","status":"PERSISTENT_VOLUME_STATUS_DELETED"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("Volume", []string{"PERSISTENT_VOLUME_STATUS_DELETED", "PERSISTENT_VOLUME_STATUS_DELETING"}, time.Minute),
		volumeStatusPoller(context.Background(), client, "vol-uuid"),
	)
	if err != nil {
		t.Fatalf("expected the DELETED volume to satisfy the wait, got %s", err)
	}
}

func TestWaitForStatusVolumeTimesOutWhenStuck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"volume":{"id":"vol-uuid","status":"PERSISTENT_VOLUME_STATUS_READY"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("Volume", []string{"PERSISTENT_VOLUME_STATUS_DELETED", "PERSISTENT_VOLUME_STATUS_DELETING"}, 0),
		volumeStatusPoller(context.Background(), client, "vol-uuid"),
	)
	if err == nil || !strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("expected a timeout error for a stuck volume, got %v", err)
	}
}

func TestWaitForStatusServicePoolSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","status":"DELETING"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("ServicePool", []string{"DELETING"}, time.Minute),
		servicePoolStatusPoller(context.Background(), client, "pool-uuid"),
	)
	if err != nil {
		t.Fatalf("expected the DELETING pool to satisfy the wait, got %s", err)
	}
}

func TestWaitForStatusServicePoolTreats404AsGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("ServicePool", []string{"DELETING"}, time.Minute),
		servicePoolStatusPoller(context.Background(), client, "pool-uuid"),
	)
	if err != nil {
		t.Fatalf("expected a 404 to count as destroyed, got %s", err)
	}
}

func TestWaitForStatusServicePoolTimesOutWhenStuck(t *testing.T) {
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
	err := waitForStatus(
		context.Background(),
		goneWait("ServicePool", []string{"DELETING"}, 0),
		servicePoolStatusPoller(context.Background(), client, "pool-uuid"),
	)
	if err == nil || !strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("expected a timeout error for a stuck pool, got %v", err)
	}
}

func TestWaitForStatusSnapshotSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"snapshot":{"id":"snap-uuid","status":"SNAPSHOT_STATUS_DELETED"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("Snapshot", []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}, time.Minute),
		snapshotStatusPoller(context.Background(), client, "snap-uuid"),
	)
	if err != nil {
		t.Fatalf("expected the DELETED snapshot to satisfy the wait, got %s", err)
	}
}

func TestWaitForStatusSnapshotTimesOutWhenStuck(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"snapshot":{"id":"snap-uuid","status":"SNAPSHOT_STATUS_CREATING"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("Snapshot", []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}, 0),
		snapshotStatusPoller(context.Background(), client, "snap-uuid"),
	)
	if err == nil || !strings.Contains(err.Error(), "timed out after") {
		t.Fatalf("expected a timeout error for a stuck snapshot, got %v", err)
	}
}

func TestWaitForStatusSnapshotTreats404AsGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("Snapshot", []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}, time.Minute),
		snapshotStatusPoller(context.Background(), client, "snap-uuid"),
	)
	if err != nil {
		t.Fatalf("expected a 404 to count as destroyed, got %s", err)
	}
}

func TestWaitForStatusHonorsContextCancellation(t *testing.T) {
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

	// The polling request carries the cancelled context: the first call
	// fails with a nil response, which must surface as an error (pinning
	// the nil-response guard) instead of a panic.
	start := time.Now()
	err := waitForStatus(
		ctx,
		goneWait("ServicePool", []string{"READY"}, time.Minute),
		servicePoolStatusPoller(ctx, client, "pool-uuid"),
	)
	if err == nil || !strings.Contains(err.Error(), "cancel") {
		t.Fatalf("expected a cancellation error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("expected the wait to unblock immediately, took %s", elapsed)
	}
}

func TestWaitForStatusSurfaces404WhenRequired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	// Readiness waits treat a mid-wait 404 as a hard error, not success.
	w := goneWait("ServicePool", []string{"READY"}, time.Minute)
	w.notFoundGone = false
	err := waitForStatus(
		context.Background(),
		w,
		servicePoolStatusPoller(context.Background(), client, "pool-uuid"),
	)
	if err == nil {
		t.Fatal("expected a mid-wait 404 to surface as an error, got nil")
	}
}

func TestWaitForStatusAppSatisfiesTargetStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"app":{"id":"app-uuid","status":"DELETED"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	err := waitForStatus(
		context.Background(),
		goneWait("App", []string{"DELETED", "DELETING"}, time.Minute),
		appStatusPoller(context.Background(), client, "app-uuid"),
	)
	if err != nil {
		t.Fatalf("expected the DELETED app to satisfy the wait, got %s", err)
	}
}

// shortenWaits collapses every wait knob so wait tests run in milliseconds;
// restore happens via t.Cleanup. Mutating these package vars requires
// tests that do not use t.Parallel().
func shortenWaits(t *testing.T) {
	t.Helper()
	originalRetry, originalService, originalPool := waitRetryInterval, serviceReadinessTimeout, servicePoolReadinessTimeout
	originalClaimTimeout, originalClaimInterval := claimWaitTimeout, claimWaitInterval
	waitRetryInterval = 5 * time.Millisecond
	serviceReadinessTimeout = 250 * time.Millisecond
	servicePoolReadinessTimeout = 250 * time.Millisecond
	claimWaitTimeout = 250 * time.Millisecond
	claimWaitInterval = 5 * time.Millisecond
	t.Cleanup(func() {
		waitRetryInterval = originalRetry
		serviceReadinessTimeout = originalService
		servicePoolReadinessTimeout = originalPool
		claimWaitTimeout = originalClaimTimeout
		claimWaitInterval = originalClaimInterval
	})
}

// firstPollThen reports first on the initial call and next afterwards.
func firstPollThen(counter *int32, first, next string) string {
	if atomic.AddInt32(counter, 1) > 1 {
		return next
	}
	return first
}

// A read whose API call fails at the transport layer returns err with a nil
// resp; the 404 check must not dereference it.
func TestResourceReadsHandleTransportErrorWithNilResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = url
	client := koyeb.NewAPIClient(cfg)

	id := "123e4567-e89b-42d3-a456-426614174000" // UUIDv4: the resolver passes it through
	reads := map[string]struct {
		resource *schema.Resource
		read     func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	}{
		"app":      {resourceKoyebApp(), resourceKoyebAppRead},
		"database": {resourceKoyebDatabase(), resourceKoyebDatabaseRead},
		"domain":   {resourceKoyebDomain(), resourceKoyebDomainRead},
		"secret":   {resourceKoyebSecret(), resourceKoyebSecretRead},
		"service":  {resourceKoyebService(), resourceKoyebServiceRead},
		"snapshot": {resourceKoyebSnapshot(), resourceKoyebSnapshotRead},
		"volume":   {resourceKoyebVolume(), resourceKoyebVolumeRead},
	}
	for name, tc := range reads {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, tc.resource.Schema, map[string]interface{}{})
			d.SetId(id)
			diags := tc.read(context.Background(), d, client)
			if !diags.HasError() {
				t.Errorf("expected an error diagnostic, got none")
			}
		})
	}
}

func TestWaitForStatusDoesNotRetryTransientErrorsWhenDisabled(t *testing.T) {
	old := waitRetryInterval
	waitRetryInterval = time.Millisecond
	defer func() { waitRetryInterval = old }()

	poll := func() (string, error) { return "", fmt.Errorf("%w: connection reset", errTransient) }
	start := time.Now()
	err := waitForStatus(context.Background(), statusWait{
		name: "svc", targets: []string{"HEALTHY"}, terminals: []string{"ERROR"},
		timeout: 5 * time.Second, interval: time.Millisecond, retryTransient: false,
	}, poll)
	if err == nil || !errors.Is(err, errTransient) {
		t.Fatalf("expected the transient error to surface, got %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatalf("wait retried a transient error although retryTransient is false")
	}
}

func TestWaitForStatusRetriesTransientErrorsWhenEnabled(t *testing.T) {
	old := waitRetryInterval
	waitRetryInterval = time.Millisecond
	defer func() { waitRetryInterval = old }()

	polls := 0
	poll := func() (string, error) {
		polls++
		return "", fmt.Errorf("%w: connection reset", errTransient)
	}
	err := waitForStatus(context.Background(), statusWait{
		name: "svc", targets: []string{"HEALTHY"}, terminals: []string{"ERROR"},
		timeout: 50 * time.Millisecond, interval: time.Millisecond, retryTransient: true,
	}, poll)
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected a timeout after retrying, got %v", err)
	}
	if polls < 2 {
		t.Fatalf("expected the transient error to be retried, got %d poll(s)", polls)
	}
}

func TestWaitForStatusAbortsNonTransientErrorsEvenWhenRetryingEnabled(t *testing.T) {
	old := waitRetryInterval
	waitRetryInterval = time.Millisecond
	defer func() { waitRetryInterval = old }()

	polls := 0
	poll := func() (string, error) {
		polls++
		return "", errors.New("boom")
	}
	err := waitForStatus(context.Background(), statusWait{
		name: "svc", targets: []string{"HEALTHY"}, terminals: []string{"ERROR"},
		timeout: 5 * time.Second, interval: time.Millisecond, retryTransient: true,
	}, poll)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected the non-transient error to abort the wait, got %v", err)
	}
	if polls != 1 {
		t.Fatalf("expected exactly one poll, got %d", polls)
	}
}

func TestPoolClaimStatusPollerPassesThroughNonGoneErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	_, err := poolClaimStatusPoller(context.Background(), client, "claim-uuid")()
	if err == nil || !errors.Is(err, errTransient) {
		t.Fatalf("expected the transient error, got %v", err)
	}
	if strings.Contains(err.Error(), "disappeared") {
		t.Fatalf("a non-gone error must not be wrapped as disappeared: %v", err)
	}
}

// A destroy retried after a crash must succeed when the resource is already
// gone (404), not error out.
func TestResourceDeletesTreat404AsGone(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	deletes := map[string]struct {
		resource *schema.Resource
		attrs    map[string]interface{}
		del      func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	}{
		"app":      {resourceKoyebApp(), nil, resourceKoyebAppDelete},
		"database": {resourceKoyebDatabase(), nil, resourceKoyebDatabaseDelete},
		"domain":   {resourceKoyebDomain(), nil, resourceKoyebDomainDelete},
		"secret":   {resourceKoyebSecret(), nil, resourceKoyebSecretDelete},
		"service":  {resourceKoyebService(), nil, resourceKoyebServiceDelete},
		"pool":     {resourceKoyebServicePool(), nil, resourceKoyebServicePoolDelete},
		"claim":    {resourceKoyebServicePoolClaim(), map[string]interface{}{"service_id": "svc-uuid"}, resourceKoyebServicePoolClaimDelete},
		"snapshot": {resourceKoyebSnapshot(), nil, resourceKoyebSnapshotDelete},
		"volume":   {resourceKoyebVolume(), nil, resourceKoyebVolumeDelete},
	}
	for name, tc := range deletes {
		t.Run(name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, tc.resource.Schema, tc.attrs)
			d.SetId("123e4567-e89b-42d3-a456-426614174000")
			diags := tc.del(context.Background(), d, client)
			if diags.HasError() {
				t.Fatalf("expected 404 to be treated as gone, got %v", diags)
			}
			if d.Id() != "" {
				t.Errorf("expected the id to be cleared, got %q", d.Id())
			}
		})
	}
}
