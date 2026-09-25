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
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccKoyebServicePool_Basic(t *testing.T) {
	var pool koyeb.ServicePool
	poolName := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebServicePoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccKoyebServicePoolConfig_basic, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServicePoolExists("koyeb_service_pool.foo", &pool),
					resource.TestCheckResourceAttr("koyeb_service_pool.foo", "name", poolName),
					resource.TestCheckResourceAttr("koyeb_service_pool.foo", "size", "1"),
					resource.TestCheckResourceAttrSet("koyeb_service_pool.foo", "id"),
					resource.TestCheckResourceAttrSet("koyeb_service_pool.foo", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_service_pool.foo", "ready_count"),
					resource.TestCheckResourceAttrSet("koyeb_service_pool.foo", "status"),
					resource.TestCheckResourceAttrSet("koyeb_service_pool.foo", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service_pool.foo", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(testAccKoyebServicePoolConfig_updated, poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServicePoolExists("koyeb_service_pool.foo", &pool),
					resource.TestCheckResourceAttr("koyeb_service_pool.foo", "size", "2"),
				),
			},
		},
	})
}

func testAccCheckKoyebServicePoolExists(n string, pool *koyeb.ServicePool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.ServicePoolsApi.GetServicePool(context.Background(), rs.Primary.ID).Execute()
		if err != nil {
			return err
		}

		servicePool := res.GetServicePool()
		if servicePool.GetId() != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*pool = servicePool

		return nil
	}
}

func testAccCheckKoyebServicePoolDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_service_pool" {
			continue
		}

		err := waitForResourceStatus(
			context.Background(),
			client.ServicePoolsApi.GetServicePool(context.Background(), rs.Primary.ID).Execute,
			"ServicePool", []string{"DELETING"}, time.Minute, false,
		)
		if err != nil {
			return fmt.Errorf("Service pool still exists: %s", err)
		}
	}

	return nil
}

const testAccKoyebServicePoolConfig_basic = `
resource "koyeb_service_pool" "foo" {
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
}`

const testAccKoyebServicePoolConfig_updated = `
resource "koyeb_service_pool" "foo" {
	name = "%s"
	size = 2
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
}`

func TestResourceKoyebServicePoolRenameReturnsNotSupportedError(t *testing.T) {
	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "renamed-pool",
		"size": 1,
	})
	d.SetId("pool-id")

	diags := resourceKoyebServicePoolUpdate(context.Background(), d, nil)

	if len(diags) != 1 {
		t.Fatalf("expected exactly 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Severity != diag.Error {
		t.Errorf("expected an error diagnostic, got severity %v", diags[0].Severity)
	}
	if !strings.Contains(diags[0].Summary, "Renaming service pools is not supported") {
		t.Errorf("expected an explicit rename-not-supported error, got: %s", diags[0].Summary)
	}
	if got := d.Id(); got != "pool-id" {
		t.Errorf("expected the pool ID to be preserved, got %q", got)
	}
}

func TestResourceKoyebServicePoolDeleteCallsAPIAndClearsID(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
	})
	d.SetId("pool-id")

	diags := resourceKoyebServicePoolDelete(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected a DELETE request, got %s", gotMethod)
	}
	if gotPath != "/v1/service_pools/pool-id" {
		t.Errorf("expected the request path /v1/service_pools/pool-id, got %s", gotPath)
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected the pool ID to be cleared, got %q", got)
	}
}

func TestSetServicePoolAttributeMapsServicePoolToState(t *testing.T) {
	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "stale-seed-pool",
		"size": 1,
	})

	pool := koyeb.ServicePool{
		Id:             toOpt("pool-123"),
		Name:           toOpt("my-pool"),
		Size:           toOpt(int64(2)),
		ReadyCount:     toOpt(int64(1)),
		OrganizationId: toOpt("org-123"),
		WorkspaceId:    toOpt("ws-123"),
		CustomerId:     toOpt("cust-123"),
		Generation:     toOpt("7"),
		Status:         toOpt(koyeb.SERVICEPOOLSTATUS_READY),
		Messages:       []string{"provisioned", "scaled"},
		CreatedAt:      toOpt(time.Date(2026, time.September, 17, 7, 51, 3, 0, time.UTC)),
		UpdatedAt:      toOpt(time.Date(2026, time.September, 17, 8, 30, 0, 0, time.UTC)),
		Definition: &koyeb.DeploymentDefinition{
			Name: toOpt("pool-def"),
			Type: toOpt(koyeb.DeploymentDefinitionType("WEB")),
			Docker: &koyeb.DockerSource{
				Image: toOpt("koyeb/demo"),
			},
			InstanceTypes: []koyeb.DeploymentInstanceType{
				{Type: toOpt("micro")},
			},
			Regions: []string{"fra"},
		},
	}

	if err := setServicePoolAttribute(d, pool); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got := d.Id(); got != "pool-123" {
		t.Errorf("unexpected ID: %q", got)
	}
	if got := d.Get("name"); got != "my-pool" {
		t.Errorf("unexpected name: %v", got)
	}
	if got := d.Get("size"); got != 2 {
		t.Errorf("unexpected size: %v", got)
	}
	if got := d.Get("ready_count"); got != 1 {
		t.Errorf("unexpected ready_count: %v", got)
	}
	if got := d.Get("status").(string); got != "READY" {
		t.Errorf("unexpected status: %q", got)
	}
	if got := d.Get("messages"); got != "provisioned scaled" {
		t.Errorf("unexpected messages: %v", got)
	}
	if got := d.Get("generation"); got != "7" {
		t.Errorf("unexpected generation: %v", got)
	}
	if got := d.Get("organization_id"); got != "org-123" {
		t.Errorf("unexpected organization_id: %v", got)
	}
	if got := d.Get("workspace_id"); got != "ws-123" {
		t.Errorf("unexpected workspace_id: %v", got)
	}
	if got := d.Get("customer_id"); got != "cust-123" {
		t.Errorf("unexpected customer_id: %v", got)
	}
	if got := d.Get("created_at"); got != "2026-09-17 07:51:03 +0000 UTC" {
		t.Errorf("unexpected created_at: %v", got)
	}
	if got := d.Get("updated_at"); got != "2026-09-17 08:30:00 +0000 UTC" {
		t.Errorf("unexpected updated_at: %v", got)
	}

	defList := d.Get("definition").([]interface{})
	if len(defList) != 1 {
		t.Fatalf("expected 1 definition block in state, got %d", len(defList))
	}
	definition := defList[0].(map[string]interface{})
	if got := definition["name"]; got != "pool-def" {
		t.Errorf("unexpected definition name: %v", got)
	}
	dockerList := definition["docker"].(*schema.Set).List()
	if len(dockerList) != 1 {
		t.Fatalf("expected 1 docker block in the definition, got %d", len(dockerList))
	}
	if got := dockerList[0].(map[string]interface{})["image"]; got != "koyeb/demo" {
		t.Errorf("unexpected docker image: %v", got)
	}
	instanceTypeList := definition["instance_types"].(*schema.Set).List()
	if len(instanceTypeList) != 1 {
		t.Fatalf("expected 1 instance_types block in the definition, got %d", len(instanceTypeList))
	}
	if got := instanceTypeList[0].(map[string]interface{})["type"]; got != "micro" {
		t.Errorf("unexpected instance type: %v", got)
	}
	if got := definition["regions"].(*schema.Set).List(); len(got) != 1 || got[0] != "fra" {
		t.Errorf("unexpected regions: %v", got)
	}
}

func TestResourceKoyebServicePoolCreateCallsAPIWithBody(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /v1/service_pools":
			gotMethod, gotPath = r.Method, r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","name":"my-pool"}}`))
		case "GET /v1/service_pools/pool-uuid":
			_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","name":"my-pool","size":1,` +
				`"ready_count":1,"status":"READY","organization_id":"org-id","workspace_id":"ws-id",` +
				`"generation":"1","messages":[],"created_at":"2026-09-18T13:00:00Z","updated_at":"2026-09-18T13:00:00Z"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
		"definition": []interface{}{
			map[string]interface{}{
				"name": "pool",
				"docker": []interface{}{
					map[string]interface{}{"image": "koyeb/demo"},
				},
			},
		},
	})

	diags := resourceKoyebServicePoolCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/service_pools" {
		t.Fatalf("expected a POST /v1/service_pools, got %s %s", gotMethod, gotPath)
	}
	if gotBody["name"] != "my-pool" || gotBody["size"] != float64(1) {
		t.Errorf("expected the body to carry the name and size, got %v", gotBody)
	}
	definition, ok := gotBody["definition"].(map[string]interface{})
	if !ok || definition["name"] != "pool" {
		t.Errorf("expected the definition name in the body, got %v", gotBody["definition"])
	}
	docker, ok := definition["docker"].(map[string]interface{})
	if !ok || docker["image"] != "koyeb/demo" {
		t.Errorf("expected the docker image in the body, got %v", definition["docker"])
	}
	if got := d.Id(); got != "pool-uuid" {
		t.Errorf("expected the pool ID to be set, got %q", got)
	}
	if got := d.Get("ready_count").(int); got != 1 {
		t.Errorf("expected the read to map the pool, got ready_count %d", got)
	}
}

func TestResourceKoyebServicePoolUpdatePutsBodyWithoutMask(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "PUT /v1/service_pools/pool-uuid":
			gotMethod, gotPath, gotQuery = r.Method, r.URL.Path, r.URL.RawQuery
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","name":"my-pool","size":2}}`))
		case "GET /v1/service_pools/pool-uuid":
			_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","name":"my-pool","size":2,` +
				`"ready_count":2,"status":"READY","organization_id":"org-id","workspace_id":"ws-id",` +
				`"generation":"1","messages":[],"created_at":"2026-09-18T13:00:00Z","updated_at":"2026-09-18T13:00:00Z"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	state := &terraform.InstanceState{
		ID: "pool-uuid",
		Attributes: map[string]string{
			"name":                        "my-pool",
			"size":                        "2",
			"definition.#":                "1",
			"definition.0.name":           "pool",
			"definition.0.docker.#":       "1",
			"definition.0.docker.0.image": "koyeb/demo",
		},
	}
	d := resourceKoyebServicePool().Data(state)

	diags := resourceKoyebServicePoolUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if gotMethod != http.MethodPut || gotPath != "/v1/service_pools/pool-uuid" {
		t.Fatalf("expected a PUT /v1/service_pools/pool-uuid, got %s %s", gotMethod, gotPath)
	}
	// The server rejects masks on PUT: the update must send a full body and
	// never an update_mask query.
	if strings.Contains(gotQuery, "update_mask") {
		t.Errorf("expected no update_mask in the query, got %q", gotQuery)
	}
	if gotBody["size"] != float64(2) {
		t.Errorf("expected the body to carry the size, got %v", gotBody)
	}
	if _, ok := gotBody["name"]; ok {
		t.Errorf("expected no name in the update body, got %v", gotBody["name"])
	}
	definition, ok := gotBody["definition"].(map[string]interface{})
	if !ok || definition["name"] != "pool" {
		t.Errorf("expected the definition in the body, got %v", gotBody["definition"])
	}
	if got := d.Get("ready_count").(int); got != 2 {
		t.Errorf("expected the read to map the pool, got ready_count %d", got)
	}
}

func TestResourceKoyebServicePoolCreateErrorsOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
		"definition": []interface{}{
			map[string]interface{}{
				"name": "pool",
				"docker": []interface{}{
					map[string]interface{}{"image": "koyeb/demo"},
				},
			},
		},
	})

	diags := resourceKoyebServicePoolCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error creating service pool") {
		t.Errorf("expected a create-failure error, got: %s", diags[0].Summary)
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected no pool ID after a failed create, got %q", got)
	}
}

func TestResourceKoyebServicePoolReadClearsIDOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
	})
	d.SetId("pool-uuid")

	diags := resourceKoyebServicePoolRead(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for a gone pool, got %v", diags)
	}
	if got := d.Id(); got != "" {
		t.Errorf("expected the pool ID to be cleared, got %q", got)
	}
}

func TestResourceKoyebServicePoolReadErrorsOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
	})
	d.SetId("pool-uuid")

	diags := resourceKoyebServicePoolRead(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error retrieving service pool") {
		t.Errorf("expected a read-failure error, got: %s", diags[0].Summary)
	}
	if got := d.Id(); got != "pool-uuid" {
		t.Errorf("expected the pool ID to be preserved, got %q", got)
	}
}

func TestResourceKoyebServicePoolUpdateErrorsOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	state := &terraform.InstanceState{
		ID: "pool-uuid",
		Attributes: map[string]string{
			"name":                        "my-pool",
			"size":                        "2",
			"definition.#":                "1",
			"definition.0.name":           "pool",
			"definition.0.docker.#":       "1",
			"definition.0.docker.0.image": "koyeb/demo",
		},
	}
	d := resourceKoyebServicePool().Data(state)

	diags := resourceKoyebServicePoolUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error updating service pool") {
		t.Errorf("expected an update-failure error, got: %s", diags[0].Summary)
	}
}

func TestResourceKoyebServicePoolDeleteErrorsOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
	})
	d.SetId("pool-uuid")

	diags := resourceKoyebServicePoolDelete(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error deleting service pool") {
		t.Errorf("expected a delete-failure error, got: %s", diags[0].Summary)
	}
	if got := d.Id(); got != "pool-uuid" {
		t.Errorf("expected the pool ID to be preserved after a failed delete, got %q", got)
	}
}

// shortenPoolWaits collapses the readiness poll interval and timeout so
// wait tests run in milliseconds; restore happens via t.Cleanup. Mutating
// these package vars requires tests that do not use t.Parallel().
func shortenPoolWaits(t *testing.T) {
	t.Helper()
	originalInterval := waitRetryInterval
	originalTimeout := servicePoolReadinessTimeout
	waitRetryInterval = 5 * time.Millisecond
	servicePoolReadinessTimeout = 250 * time.Millisecond
	t.Cleanup(func() {
		waitRetryInterval = originalInterval
		servicePoolReadinessTimeout = originalTimeout
	})
}

// poolWaitTestServer serves a create/update reply and a GetServicePool that
// reports PROVISIONING on the first poll and the given status afterwards.
func poolWaitTestServer(t *testing.T, finalStatus string) (*httptest.Server, *int32) {
	t.Helper()
	var gets int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /v1/service_pools", "PUT /v1/service_pools/pool-uuid":
			_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","name":"my-pool","size":1}}`))
		case "GET /v1/service_pools/pool-uuid":
			status := "PROVISIONING"
			if n := atomic.AddInt32(&gets, 1); n > 1 {
				status = finalStatus
			}
			_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","name":"my-pool","size":1,` +
				`"ready_count":1,"status":"` + status + `","organization_id":"org-id","workspace_id":"ws-id",` +
				`"generation":"1","messages":[],"created_at":"2026-09-18T13:00:00Z","updated_at":"2026-09-18T13:00:00Z"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, &gets
}

func TestResourceKoyebServicePoolCreateWaitsForPoolReady(t *testing.T) {
	shortenPoolWaits(t)
	srv, gets := poolWaitTestServer(t, "READY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
		"definition": []interface{}{
			map[string]interface{}{
				"name": "pool",
				"docker": []interface{}{
					map[string]interface{}{"image": "koyeb/demo"},
				},
			},
		},
	})

	diags := resourceKoyebServicePoolCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Get("status").(string); got != "READY" {
		t.Errorf("expected the create to return a READY pool, got %q", got)
	}
	if n := atomic.LoadInt32(gets); n < 2 {
		t.Errorf("expected the create to poll GetServicePool at least twice, got %d polls", n)
	}
}

func TestResourceKoyebServicePoolCreateFailsWhenPoolNeverReady(t *testing.T) {
	shortenPoolWaits(t)
	srv, _ := poolWaitTestServer(t, "PROVISIONING")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
		"definition": []interface{}{
			map[string]interface{}{
				"name": "pool",
				"docker": []interface{}{
					map[string]interface{}{"image": "koyeb/demo"},
				},
			},
		},
	})

	diags := resourceKoyebServicePoolCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error waiting for service pool") {
		t.Errorf("expected a wait-failure error, got: %s", diags[0].Summary)
	}
}

func TestResourceKoyebServicePoolUpdateWaitsForPoolReady(t *testing.T) {
	shortenPoolWaits(t)
	srv, gets := poolWaitTestServer(t, "READY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	state := &terraform.InstanceState{
		ID: "pool-uuid",
		Attributes: map[string]string{
			"name":                        "my-pool",
			"size":                        "1",
			"definition.#":                "1",
			"definition.0.name":           "pool",
			"definition.0.docker.#":       "1",
			"definition.0.docker.0.image": "koyeb/demo",
		},
	}
	d := resourceKoyebServicePool().Data(state)

	diags := resourceKoyebServicePoolUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Get("status").(string); got != "READY" {
		t.Errorf("expected the update to return a READY pool, got %q", got)
	}
	if n := atomic.LoadInt32(gets); n < 2 {
		t.Errorf("expected the update to poll GetServicePool at least twice, got %d polls", n)
	}
}

func TestResourceKoyebServicePoolCreateFailsFastWhenPoolErrors(t *testing.T) {
	shortenPoolWaits(t)
	srv, gets := poolWaitTestServer(t, "ERROR")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, servicePoolSchema(), map[string]interface{}{
		"name": "my-pool",
		"size": 1,
		"definition": []interface{}{
			map[string]interface{}{
				"name": "pool",
				"docker": []interface{}{
					map[string]interface{}{"image": "koyeb/demo"},
				},
			},
		},
	})

	diags := resourceKoyebServicePoolCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "ERROR") {
		t.Errorf("expected the diagnostic to surface the ERROR pool status, got: %s", diags[0].Summary)
	}
	if n := atomic.LoadInt32(gets); n != 2 {
		t.Errorf("expected the wait to stop at the first ERROR poll, got %d polls", n)
	}
}
