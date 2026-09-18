package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
