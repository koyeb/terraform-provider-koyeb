package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccDataSourceKoyebServicePool_Basic(t *testing.T) {
	var pool koyeb.ServicePool
	poolName := randomTestName()

	resourceConfig := fmt.Sprintf(testAccKoyebServicePoolConfig_basic, poolName)

	dataSourceConfig := `
data "koyeb_service_pool" "bar" {
	name = koyeb_service_pool.foo.name
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccSkipIfServicePoolsUnavailable(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
			},
			{
				Config: resourceConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceKoyebServicePoolExists("data.koyeb_service_pool.bar", &pool),
					resource.TestCheckResourceAttr("data.koyeb_service_pool.bar", "name", poolName),
					resource.TestCheckResourceAttrSet("data.koyeb_service_pool.bar", "id"),
					resource.TestCheckResourceAttrSet("data.koyeb_service_pool.bar", "size"),
					resource.TestCheckResourceAttrSet("data.koyeb_service_pool.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_service_pool.bar", "ready_count"),
					resource.TestCheckResourceAttrSet("data.koyeb_service_pool.bar", "status"),
				),
			},
		},
	})
}

func testAccCheckDataSourceKoyebServicePoolExists(n string, pool *koyeb.ServicePool) resource.TestCheckFunc {
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

func TestDataSourceKoyebServicePoolReadResolvesPoolByName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/service_pools":
			_, _ = w.Write([]byte(`{"service_pools":[{"id":"pool-uuid","name":"my-pool"}]}`))
		case "/v1/service_pools/pool-uuid":
			_, _ = w.Write([]byte(`{"service_pool":{"id":"pool-uuid","name":"my-pool","size":2,` +
				`"ready_count":1,"status":"READY","organization_id":"org-id","workspace_id":"ws-id",` +
				`"generation":"1","messages":[],"created_at":"2026-09-18T13:00:00Z","updated_at":"2026-09-18T13:00:00Z"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, dataSourceKoyebServicePool().Schema, map[string]interface{}{
		"name": "my-pool",
	})

	diags := dataSourceKoyebServicePoolRead(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Id(); got != "pool-uuid" {
		t.Errorf("expected the pool ID, got %q", got)
	}
	if got := d.Get("name").(string); got != "my-pool" {
		t.Errorf("unexpected name: %q", got)
	}
	if got := d.Get("size").(int); got != 2 {
		t.Errorf("unexpected size: %d", got)
	}
	if got := d.Get("status").(string); got != "READY" {
		t.Errorf("unexpected status: %q", got)
	}
}

func TestDataSourceKoyebServicePoolReadErrorsOnListFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, dataSourceKoyebServicePool().Schema, map[string]interface{}{
		"name": "my-pool",
	})

	diags := dataSourceKoyebServicePoolRead(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error retrieving service pool") {
		t.Errorf("expected a list-failure error, got: %s", diags[0].Summary)
	}
}

func TestDataSourceKoyebServicePoolReadErrorsOnUnknownPool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"service_pools":[{"id":"other-uuid","name":"other"}]}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, dataSourceKoyebServicePool().Schema, map[string]interface{}{
		"name": "my-pool",
	})

	diags := dataSourceKoyebServicePoolRead(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "No service pool found with name my-pool") {
		t.Errorf("expected an unknown-pool error, got: %s", diags[0].Summary)
	}
}
