package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestDatabaseSchemaContract(t *testing.T) {
	s := databaseSchema()

	name := s["name"]
	if !name.Required {
		t.Error("expected name to be required")
	}
	if !name.ForceNew {
		t.Error("expected name to force new resources")
	}

	for _, key := range []string{"pg_version", "region", "instance_type", "db_name", "db_owner"} {
		field, ok := s[key]
		if !ok {
			t.Errorf("expected %q in the database schema", key)
			continue
		}
		if !field.Optional {
			t.Errorf("expected %q to be optional", key)
		}
	}

	defaults := map[string]interface{}{
		"pg_version":    16,
		"region":        "was",
		"instance_type": "free",
		"db_name":       "koyebdb",
		"db_owner":      "koyeb-adm",
	}
	for key, want := range defaults {
		field, ok := s[key]
		if !ok {
			t.Errorf("expected %q in the database schema", key)
			continue
		}
		if field.Default != want {
			t.Errorf("expected %q default %v, got %v", key, want, field.Default)
		}
	}

	for _, key := range []string{"id", "organization_id", "app_id", "role_secret", "status", "version", "updated_at", "created_at"} {
		field, ok := s[key]
		if !ok {
			t.Errorf("expected %q in the database schema", key)
			continue
		}
		if !field.Computed {
			t.Errorf("expected %q to be computed", key)
		}
	}
}

func TestBuildDatabaseDefinition(t *testing.T) {
	definition := buildDatabaseDefinition("my-db", 16, "was", "free", "koyebdb", "koyeb-adm", "role-secret")

	if definition.GetName() != "my-db" {
		t.Errorf("expected name my-db, got %q", definition.GetName())
	}
	if definition.GetType() != koyeb.DEPLOYMENTDEFINITIONTYPE_DATABASE {
		t.Errorf("expected type %q, got %q", koyeb.DEPLOYMENTDEFINITIONTYPE_DATABASE, definition.GetType())
	}

	if definition.Database == nil || definition.Database.NeonPostgres == nil {
		t.Fatal("expected a neon postgres database source")
	}

	neon := definition.Database.NeonPostgres
	if neon.GetPgVersion() != 16 {
		t.Errorf("expected pg_version 16, got %d", neon.GetPgVersion())
	}
	if neon.GetRegion() != "was" {
		t.Errorf("expected region was, got %q", neon.GetRegion())
	}
	if neon.GetInstanceType() != "free" {
		t.Errorf("expected instance type free, got %q", neon.GetInstanceType())
	}
	if len(neon.Databases) != 1 || neon.Databases[0].GetName() != "koyebdb" || neon.Databases[0].GetOwner() != "koyeb-adm" {
		t.Errorf("expected one database koyebdb owned by koyeb-adm, got %+v", neon.Databases)
	}
	if len(neon.Roles) != 1 || neon.Roles[0].GetName() != "koyeb-adm" || neon.Roles[0].GetSecret() != "role-secret" {
		t.Errorf("expected one role koyeb-adm with secret role-secret, got %+v", neon.Roles)
	}
}

func TestDatabaseDataSourceSchemaContract(t *testing.T) {
	s := databaseDataSourceSchema()

	name := s["name"]
	if !name.Required {
		t.Error("expected name to be required to look up a database")
	}

	for _, key := range []string{"id", "app_id", "organization_id", "status", "version", "updated_at", "created_at"} {
		field, ok := s[key]
		if !ok {
			t.Errorf("expected %q in the database datasource schema", key)
			continue
		}
		if !field.Computed {
			t.Errorf("expected %q to be computed", key)
		}
	}

	if _, ok := s["role_secret"]; ok {
		t.Error("expected role_secret to be absent from the datasource: it lives in the service definition, not the service object")
	}
}

func TestAccKoyebDatabase_Basic(t *testing.T) {
	// The free Neon instance quota of the CI organization fits a single
	// database, and the Terraform matrix runs this suite on three versions
	// in parallel; run on one version only.
	if v := os.Getenv("TF_ACC_MATRIX"); v != "" && !strings.HasPrefix(v, "1.1") {
		t.Skipf("skipping to stay within the organization's free instance quota (TF %s)", v)
	}

	// Skip when the organization's free instance quota is already
	// consumed: that is an environment limit, not a code defect.
	checkFreeQuota := func() {
		if exhausted, err := freeInstanceQuotaExhausted(testAccProvider.Meta().(*koyeb.APIClient)); err == nil && exhausted {
			t.Skip("skipping: the organization's free instance quota is exhausted")
		}
	}

	var service koyeb.Service
	databaseName := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			checkFreeQuota()
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebDatabaseConfig_basic, databaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebDatabaseExists("koyeb_database.foobar", &service),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "name", databaseName),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "pg_version", "16"),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "region", "was"),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "instance_type", "free"),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "db_name", "koyebdb"),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "db_owner", "koyeb-adm"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "app_id"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "role_secret"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "version"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_database.foobar", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebDatabaseConfig_update, databaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebDatabaseExists("koyeb_database.foobar", &service),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "name", databaseName),
					resource.TestCheckResourceAttr("koyeb_database.foobar", "db_name", "koyebdb2"),
				),
			},
			{
				// The data source shares the same database: a second one would
				// race the organization's single free instance quota.
				Config: fmt.Sprintf(testAccCheckKoyebDatabaseConfig_update, databaseName) + `
data "koyeb_database" "bar" {
	name = koyeb_database.foobar.name
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.koyeb_database.bar", "name", databaseName),
					resource.TestCheckResourceAttrSet("data.koyeb_database.bar", "id"),
					resource.TestCheckResourceAttrSet("data.koyeb_database.bar", "app_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_database.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_database.bar", "status"),
					resource.TestCheckResourceAttrSet("data.koyeb_database.bar", "version"),
					resource.TestCheckResourceAttrSet("data.koyeb_database.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_database.bar", "created_at"),
				),
			},
		},
	})
}

func testAccCheckKoyebDatabaseDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"DELETED", "DELETING"}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_database" {
			continue
		}

		err := waitForResourceStatus(context.Background(), client.ServicesApi.GetService(context.Background(), rs.Primary.ID).Execute, "Service", targetStatus, time.Minute, false)
		if err != nil {
			return fmt.Errorf("Database still exists: %s", err)
		}
	}

	return nil
}

func testAccCheckKoyebDatabaseExists(n string, service *koyeb.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.ServicesApi.GetService(context.Background(), rs.Primary.ID).Execute()
		if err != nil {
			return err
		}

		if *res.Service.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*service = *res.Service

		return nil
	}
}

const testAccCheckKoyebDatabaseConfig_basic = `
resource "koyeb_database" "foobar" {
	name = "%s"
}`

const testAccCheckKoyebDatabaseConfig_update = `
resource "koyeb_database" "foobar" {
	name    = "%s"
	db_name = "koyebdb2"
}`

func TestResourceKoyebDatabaseUpdateRefusesUnknownRoleSecret(t *testing.T) {
	// A database imported without a role_secret in state must not be
	// updated: the API would silently redefine the Neon role binding.
	d := schema.TestResourceDataRaw(t, databaseSchema(), map[string]interface{}{
		"name": "my-db",
	})
	d.SetId("d290f1ee-6c54-4b01-90e6-d7015f3f7b1f")

	diags := resourceKoyebDatabaseUpdate(context.Background(), d, nil)
	if !diags.HasError() {
		t.Fatal("expected an error when role_secret is unknown")
	}
	if !strings.Contains(diags[0].Summary, "role_secret") {
		t.Errorf("expected the error to mention role_secret, got %q", diags[0].Summary)
	}
}

func TestIsAppNameAlreadyExistsError(t *testing.T) {
	// The API client only exposes GenericOpenAPIError's model after
	// parsing a response, so drive CreateApp against a mock to build it.
	createAppErr := func(t *testing.T, fields string) error {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message": "invalid request", "fields": ` + fields + `}`))
		}))
		t.Cleanup(srv.Close)

		cfg := koyeb.NewConfiguration()
		cfg.Servers[0].URL = srv.URL
		client := koyeb.NewAPIClient(cfg)

		_, _, err := client.AppsApi.CreateApp(context.Background()).App(koyeb.CreateApp{Name: toOpt("my-app")}).Execute()
		if err == nil {
			t.Fatal("expected the mocked CreateApp to fail")
		}
		return err
	}

	tests := []struct {
		name   string
		fields string
		want   bool
	}{
		{
			name:   "name already exists",
			fields: `[{"field": "name", "description": "already exists"}]`,
			want:   true,
		},
		{
			name:   "reworded already-exists message",
			fields: `[{"field": "name", "description": "An app with this name already exists."}]`,
			want:   true,
		},
		{
			name:   "different field",
			fields: `[{"field": "organization", "description": "already exists"}]`,
			want:   false,
		},
		{
			name:   "unrelated description",
			fields: `[{"field": "name", "description": "is invalid"}]`,
			want:   false,
		},
		{
			name:   "multiple fields",
			fields: `[{"field": "name", "description": "already exists"}, {"field": "other", "description": "is invalid"}]`,
			want:   false,
		},
		{
			name:   "no fields",
			fields: `[]`,
			want:   false,
		},
		{
			name:   "mixed-case already-exists message",
			fields: `[{"field": "name", "description": "An app with this name Already Exists."}]`,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createAppErr(t, tt.fields)
			if got := isAppNameAlreadyExistsError(err); got != tt.want {
				t.Errorf("isAppNameAlreadyExistsError() = %v, want %v", got, tt.want)
			}
		})
	}

	if isAppNameAlreadyExistsError(fmt.Errorf("something else")) {
		t.Error("expected a plain error not to match")
	}
}

func TestResourceKoyebDatabaseReadResolvesShortID(t *testing.T) {
	const appID = "d290f1ee-6c54-4b01-90e6-d7015f3f7b1f"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/apps") && r.Method == "GET":
			// the service mapper resolves app names for its compound keys
			_, _ = w.Write([]byte(`{"apps":[{"id":"f19f2eaf-6d64-4a1c-b1d0-7015f3f7b1a","name":"my-app"}],"count":1}`))
		case strings.HasSuffix(r.URL.Path, "/services") && r.Method == "GET":
			// idmapper fetch: resolve the short ID to the service ID
			_, _ = w.Write([]byte(`{"services":[{"id":"` + appID + `","name":"my-db","app_id":"f19f2eaf-6d64-4a1c-b1d0-7015f3f7b1a"}],"count":1}`))
		case strings.Contains(r.URL.Path, appID):
			_, _ = w.Write([]byte(`{"service":{"id":"` + appID + `","name":"my-db"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	d := schema.TestResourceDataRaw(t, databaseSchema(), map[string]interface{}{
		"name": "my-db",
	})
	// A short service ID, as produced by `terraform import` with the
	// 8-character form.
	d.SetId(appID[:8])

	if diags := resourceKoyebDatabaseRead(context.Background(), d, client); diags.HasError() {
		t.Fatalf("expected no error, got %v", diags)
	}
	if d.Id() != appID {
		t.Errorf("expected the short ID to resolve to %q, got %q", appID, d.Id())
	}
}

func TestFreeInstanceQuotaExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/organizations"):
			_, _ = w.Write([]byte(`{"organizations":[{"id":"org-id","name":"test"}]}`))
		case strings.HasSuffix(r.URL.Path, "/usage"):
			_, _ = w.Write([]byte(`{"usage":{"instances_by_type":[{"instance_type":"free","used":1,"limit":1}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	exhausted, err := freeInstanceQuotaExhausted(client)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if !exhausted {
		t.Error("expected the free instance quota to be exhausted")
	}
}

func TestFreeInstanceQuotaAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/organizations"):
			_, _ = w.Write([]byte(`{"organizations":[{"id":"org-id","name":"test"}]}`))
		case strings.HasSuffix(r.URL.Path, "/usage"):
			_, _ = w.Write([]byte(`{"usage":{"instances_by_type":[{"instance_type":"free","used":0,"limit":1}]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	exhausted, err := freeInstanceQuotaExhausted(client)
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}
	if exhausted {
		t.Error("expected the free instance quota to be available")
	}
}

func TestNewRoleSecretNameIsValidSecretName(t *testing.T) {
	// The API requires secret names to match ^[a-zA-Z_][a-zA-Z0-9-_]*$
	// (2-64 chars); a bare UUID fails ~62% of the time because hex
	// UUIDs often start with a digit.
	name, err := newRoleSecretName()
	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9-_]{1,63}$`).MatchString(name) {
		t.Errorf("expected a valid secret name, got %q", name)
	}
	if !strings.HasPrefix(name, "role-") {
		t.Errorf("expected the name to carry a letter prefix, got %q", name)
	}
}

// databaseWaitTestServer serves the app create/list, the service
// create/update, and a GetService that reports STARTING on the first poll
// and the given status afterwards.
func databaseWaitTestServer(t *testing.T, finalStatus string) (*httptest.Server, *int32) {
	t.Helper()
	var gets int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /v1/apps":
			_, _ = w.Write([]byte(`{}`))
		case "PUT /v1/services/" + testServiceUUID:
			_, _ = w.Write([]byte(`{"service":{"id":"` + testServiceUUID + `","name":"my-db"}}`))
		case "GET /v1/apps":
			_, _ = w.Write([]byte(`{"apps":[{"id":"123e4567-e89b-42d3-a456-426614174000","name":"my-db"}]}`))
		case "POST /v1/services":
			_, _ = w.Write([]byte(`{"service":{"id":"` + testServiceUUID + `","name":"my-db"}}`))
		case "GET /v1/services/" + testServiceUUID:
			status := "STARTING"
			if n := atomic.AddInt32(&gets, 1); n > 1 {
				status = finalStatus
			}
			_, _ = w.Write([]byte(`{"service":{"id":"` + testServiceUUID + `","name":"my-db","status":"` + status + `"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, &gets
}

func TestResourceKoyebDatabaseCreateWaitsForDatabaseHealth(t *testing.T) {
	shortenServiceWaits(t)
	srv, gets := databaseWaitTestServer(t, "HEALTHY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, databaseSchema(), map[string]interface{}{
		"name": "my-db",
	})

	diags := resourceKoyebDatabaseCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Get("status").(string); got != "HEALTHY" {
		t.Errorf("expected the create to return a HEALTHY database, got %q", got)
	}
	if n := atomic.LoadInt32(gets); n < 2 {
		t.Errorf("expected the create to poll GetService at least twice, got %d polls", n)
	}
}

func TestResourceKoyebDatabaseUpdateWaitsForDatabaseHealth(t *testing.T) {
	shortenServiceWaits(t)
	srv, gets := databaseWaitTestServer(t, "HEALTHY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	state := &terraform.InstanceState{
		ID: testServiceUUID,
		Attributes: map[string]string{
			"name":        "my-db",
			"role_secret": "role-existing-secret",
		},
	}
	d := resourceKoyebDatabase().Data(state)

	diags := resourceKoyebDatabaseUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Get("status").(string); got != "HEALTHY" {
		t.Errorf("expected the update to return a HEALTHY database, got %q", got)
	}
	if n := atomic.LoadInt32(gets); n < 2 {
		t.Errorf("expected the update to poll GetService at least twice, got %d polls", n)
	}
}
