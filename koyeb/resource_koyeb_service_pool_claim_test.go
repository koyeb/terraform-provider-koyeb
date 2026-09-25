package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccKoyebServicePoolClaim_Basic(t *testing.T) {
	requestID := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccKoyebServicePoolClaimConfig_basic, randomTestName(), requestID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("koyeb_service_pool_claim.example", "id"),
					resource.TestCheckResourceAttrSet("koyeb_service_pool_claim.example", "service_id"),
					resource.TestCheckResourceAttr("koyeb_service_pool_claim.example", "status", "FULFILLED"),
				),
			},
		},
	})
}

const testAccKoyebServicePoolClaimConfig_basic = `
resource "koyeb_service_pool" "example" {
	name = "%s"
	size = 1
	definition {
		name = "pool"
		instance_types {
			type = "micro"
		}
		ports {
			port     = 3000
			protocol = "http"
		}
		scalings {
			min = 1
			max = 1
		}
		regions = ["tyo"]
		docker {
			image = "koyeb/demo"
		}
	}
}

resource "koyeb_service_pool_claim" "example" {
	pool       = koyeb_service_pool.example.name
	request_id = "%s"
}
`

func TestResourceKoyebServicePoolClaimCreateClaimsAndMaps(t *testing.T) {
	var (
		gotClaimMethod, gotClaimPath string
		gotClaimBody                 map[string]string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/service_pools":
			_, _ = w.Write([]byte(`{"service_pools":[{"id":"pool-uuid","name":"my-pool"}]}`))
		case "/v1/claim":
			gotClaimMethod, gotClaimPath = r.Method, r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&gotClaimBody)
			_, _ = w.Write([]byte(`{"claim_id":"claim-uuid","service_id":"service-uuid","prewarmed":true}`))
		case "/v1/claims/claim-uuid":
			_, _ = w.Write([]byte(`{
				"claim": {
					"id": "claim-uuid",
					"pool_id": "pool-uuid",
					"service_id": "service-uuid",
					"request_id": "req-42",
					"status": "FULFILLED",
					"organization_id": "org-id",
					"workspace_id": "ws-id",
					"customer_id": "customer-id",
					"created_at": "2026-09-18T13:10:00Z",
					"fulfilled_at": "2026-09-18T13:10:01Z",
					"pool_generation": "3"
				}
			}`))
		case "/v1/services/service-uuid":
			_, _ = w.Write([]byte(`{"service":{"id":"service-uuid","status":"HEALTHY"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if gotClaimMethod != http.MethodPost || gotClaimPath != "/v1/claim" {
		t.Errorf("expected a POST /v1/claim, got %s %s", gotClaimMethod, gotClaimPath)
	}
	if gotClaimBody["pool_id"] != "pool-uuid" || gotClaimBody["request_id"] != "req-42" {
		t.Errorf("expected the claim body to target the pool and carry the request ID, got %v", gotClaimBody)
	}
	if got := d.Id(); got != "claim-uuid" {
		t.Errorf("expected the claim ID to be set, got %q", got)
	}
	for key, want := range map[string]string{
		"service_id":      "service-uuid",
		"status":          "FULFILLED",
		"pool_id":         "pool-uuid",
		"created_at":      "2026-09-18 13:10:00 +0000 UTC",
		"fulfilled_at":    "2026-09-18 13:10:01 +0000 UTC",
		"pool_generation": "3",
		"organization_id": "org-id",
		"workspace_id":    "ws-id",
		"customer_id":     "customer-id",
	} {
		if got := d.Get(key).(string); got != want {
			t.Errorf("expected %s %q, got %q", key, want, got)
		}
	}
	if got := d.Get("prewarmed").(bool); !got {
		t.Errorf("expected the claim to be prewarmed, got %v", got)
	}
}

func TestResourceKoyebServicePoolClaimDeleteDeletesClaimedService(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})
	d.SetId("claim-uuid")
	if err := d.Set("service_id", "service-uuid"); err != nil {
		t.Fatal(err)
	}

	diags := resourceKoyebServicePoolClaimDelete(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected a DELETE request, got %s", gotMethod)
	}
	if gotPath != "/v1/services/service-uuid" {
		t.Errorf("expected the claimed service to be deleted at /v1/services/service-uuid, got %s", gotPath)
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected the claim ID to be cleared, got %q", got)
	}
}

func TestResourceKoyebServicePoolClaimDeleteWithoutServiceIsNoOp(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})
	d.SetId("claim-uuid")

	diags := resourceKoyebServicePoolClaimDelete(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if len(calls) != 0 {
		t.Errorf("expected no API call for an unfulfilled claim, got %v", calls)
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected the claim ID to be cleared, got %q", got)
	}
}

func TestResourceKoyebServicePoolClaimCreateErrorsOnUnknownPool(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service_pools":[]}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "no-such-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 {
		t.Fatalf("expected exactly 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != diag.Error {
		t.Errorf("expected an error diagnostic, got severity %v", diags[0].Severity)
	}
	if !strings.Contains(diags[0].Summary, "No service pool found with name no-such-pool") {
		t.Errorf("expected an unknown-pool error, got: %s", diags[0].Summary)
	}
	for _, call := range calls {
		if strings.Contains(call, "/v1/claim") {
			t.Errorf("expected no claim call for an unknown pool, got %s", call)
		}
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected no claim ID, got %q", got)
	}
}

func TestResourceKoyebServicePoolClaimReadClearsIDOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})
	d.SetId("claim-uuid")

	diags := resourceKoyebServicePoolClaimRead(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected the claim ID to be cleared for a gone claim, got %q", got)
	}
}

func TestResourceKoyebServicePoolClaimCreateErrorsOnListFailure(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error retrieving service pool") {
		t.Errorf("expected a list-failure error, got: %s", diags[0].Summary)
	}
	for _, call := range calls {
		if strings.Contains(call, "/v1/claim") {
			t.Errorf("expected no claim call after a failed lookup, got %s", call)
		}
	}
}

func TestResourceKoyebServicePoolClaimCreateErrorsOnClaimFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/service_pools":
			_, _ = w.Write([]byte(`{"service_pools":[{"id":"pool-uuid","name":"my-pool"}]}`))
		case "/v1/claim":
			http.Error(w, "boom", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error claiming service pool") {
		t.Errorf("expected a claim-failure error, got: %s", diags[0].Summary)
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected no claim ID after a failed claim, got %q", got)
	}
}

func TestResourceKoyebServicePoolClaimReadErrorsOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})
	d.SetId("claim-uuid")

	diags := resourceKoyebServicePoolClaimRead(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error retrieving service pool claim") {
		t.Errorf("expected a read-failure error, got: %s", diags[0].Summary)
	}
	if got := d.Id(); got != "claim-uuid" {
		t.Errorf("expected the claim ID to be preserved, got %q", got)
	}
}

func TestResourceKoyebServicePoolClaimDeleteErrorsOnServiceDeletionFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})
	d.SetId("claim-uuid")
	if err := d.Set("service_id", "service-uuid"); err != nil {
		t.Fatal(err)
	}

	diags := resourceKoyebServicePoolClaimDelete(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error deleting claimed service") {
		t.Errorf("expected a delete-failure error, got: %s", diags[0].Summary)
	}
	if got := d.Id(); got != "claim-uuid" {
		t.Errorf("expected the claim ID to be preserved after a failed delete, got %q", got)
	}
}

// shortenClaimWaits collapses the claim wait and service poll intervals and
// timeouts so wait tests run in milliseconds; restore happens via t.Cleanup.
func shortenClaimWaits(t *testing.T) {
	t.Helper()
	originalTimeout := claimWaitTimeout
	originalInterval := claimWaitInterval
	originalRetry := waitRetryInterval
	claimWaitTimeout = 250 * time.Millisecond
	claimWaitInterval = 5 * time.Millisecond
	waitRetryInterval = 5 * time.Millisecond
	t.Cleanup(func() {
		claimWaitTimeout = originalTimeout
		claimWaitInterval = originalInterval
		waitRetryInterval = originalRetry
	})
}

// claimWaitTestServer serves the pool lookup, the claim call, a GetClaim
// that reports PENDING on the first poll and the given status afterwards,
// and a GetService for the claimed service that reports STARTING on the
// first poll and HEALTHY afterwards.
func claimWaitTestServer(t *testing.T, finalStatus string) (*httptest.Server, *int32, *int32) {
	t.Helper()
	var gets, serviceGets int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/service_pools":
			_, _ = w.Write([]byte(`{"service_pools":[{"id":"pool-uuid","name":"my-pool"}]}`))
		case "/v1/claim":
			_, _ = w.Write([]byte(`{"claim_id":"claim-uuid","service_id":"service-uuid","prewarmed":true}`))
		case "/v1/claims/claim-uuid":
			status := "PENDING"
			if n := atomic.AddInt32(&gets, 1); n > 1 {
				status = finalStatus
			}
			_, _ = w.Write([]byte(`{"claim":{"id":"claim-uuid","pool_id":"pool-uuid",` +
				`"service_id":"service-uuid","request_id":"req-42","status":"` + status + `",` +
				`"organization_id":"org-id","workspace_id":"ws-id","customer_id":"customer-id",` +
				`"created_at":"2026-09-18T13:10:00Z","fulfilled_at":"2026-09-18T13:10:01Z",` +
				`"pool_generation":"3"}}`))
		case "/v1/services/service-uuid":
			status := "STARTING"
			if atomic.AddInt32(&serviceGets, 1) > 1 {
				status = "HEALTHY"
			}
			_, _ = w.Write([]byte(`{"service":{"id":"service-uuid","status":"` + status + `"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, &gets, &serviceGets
}

func TestResourceKoyebServicePoolClaimCreateWaitsForFulfilled(t *testing.T) {
	shortenClaimWaits(t)
	srv, gets, _ := claimWaitTestServer(t, "FULFILLED")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Get("status").(string); got != "FULFILLED" {
		t.Errorf("expected the create to return a FULFILLED claim, got %q", got)
	}
	if n := atomic.LoadInt32(gets); n < 2 {
		t.Errorf("expected the create to poll GetClaim at least twice, got %d polls", n)
	}
}

func TestResourceKoyebServicePoolClaimCreateFailsWhenClaimFailed(t *testing.T) {
	shortenClaimWaits(t)
	srv, gets, _ := claimWaitTestServer(t, "FAILED")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "FAILED") {
		t.Errorf("expected the diagnostic to surface the FAILED claim status, got: %s", diags[0].Summary)
	}
	if n := atomic.LoadInt32(gets); n != 2 {
		t.Errorf("expected the wait to stop at the first FAILED poll, got %d polls", n)
	}
}

func TestResourceKoyebServicePoolClaimCreateTimesOutWhenPending(t *testing.T) {
	shortenClaimWaits(t)
	srv, _, _ := claimWaitTestServer(t, "PENDING")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	start := time.Now()
	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "did not reach FULFILLED") {
		t.Errorf("expected a claim wait timeout error, got: %s", diags[0].Summary)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("expected the timeout to surface quickly for tests, took %s", elapsed)
	}
}

func TestWaitForClaimFulfilledHonorsContextCancellation(t *testing.T) {
	srv, _, _ := claimWaitTestServer(t, "PENDING")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := waitForClaimFulfilled(ctx, koyeb.NewAPIClient(cfg), "claim-uuid", time.Minute, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "cancel") {
		t.Fatalf("expected a cancellation error, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("expected the wait to unblock immediately, took %s", elapsed)
	}
}

func TestResourceKoyebServicePoolClaimCreateFailsWhenClaimReleased(t *testing.T) {
	shortenClaimWaits(t)
	srv, gets, _ := claimWaitTestServer(t, "RELEASED")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "RELEASED") {
		t.Errorf("expected the diagnostic to surface the RELEASED claim status, got: %s", diags[0].Summary)
	}
	if n := atomic.LoadInt32(gets); n != 2 {
		t.Errorf("expected the wait to stop at the first RELEASED poll, got %d polls", n)
	}
}

func TestResourceKoyebServicePoolClaimCreateWaitsForClaimedService(t *testing.T) {
	shortenClaimWaits(t)
	srv, _, serviceGets := claimWaitTestServer(t, "FULFILLED")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, resourceKoyebServicePoolClaim().Schema, map[string]interface{}{
		"pool":       "my-pool",
		"request_id": "req-42",
	})

	diags := resourceKoyebServicePoolClaimCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	// FULFILLED hands out a service that may still be STARTING on the
	// cold path: the create must also wait for the service itself.
	if n := atomic.LoadInt32(serviceGets); n < 2 {
		t.Errorf("expected the create to poll GetService at least twice, got %d polls", n)
	}
	if got := d.Get("status").(string); got != "FULFILLED" {
		t.Errorf("expected a FULFILLED claim, got %q", got)
	}
}
