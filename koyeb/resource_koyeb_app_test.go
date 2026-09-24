package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func init() {
	resource.AddTestSweepers("koyeb_app", &resource.Sweeper{
		Name: "koyeb_app",
		F:    testSweepApp,
	})

}

func testSweepApp(string) error {
	meta, err := sharedConfig()
	if err != nil {
		return err
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.AppsApi.ListApps(context.Background()).Limit("100").Execute()
	if err != nil {
		return err
	}

	for _, a := range res.Apps {
		if strings.HasPrefix(a.GetName(), testNamePrefix) {
			log.Printf("Destroying app %s", *a.Name)

			if _, _, err := client.AppsApi.DeleteApp(context.Background(), a.GetId()).Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccKoyebApp_Basic(t *testing.T) {
	var app koyeb.App
	appName := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebAppConfig_basic, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebAppExists("koyeb_app.foobar", &app),
					testAccCheckKoyebAppAttributes(&app, appName),
					resource.TestCheckResourceAttr(
						"koyeb_app.foobar", "name", appName),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.id"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.app_name"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.created_at"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.deployment_group"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.name"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.status"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.type"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.created_at"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "domains.0.version"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebAppConfig_delete_when_empty, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebAppExists("koyeb_app.foobar", &app),
					resource.TestCheckResourceAttr("koyeb_app.foobar", "name", appName),
					resource.TestCheckResourceAttr("koyeb_app.foobar", "delete_when_empty", "false"),
					resource.TestCheckResourceAttrSet("koyeb_app.foobar", "id"),
				),
			},
		},
	})
}

func testAccCheckKoyebAppDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"DELETED", "DELETING"}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_app" {
			continue
		}

		err := waitForResourceStatus(client.AppsApi.GetApp(context.Background(), rs.Primary.ID).Execute, "App", targetStatus, 1, false)
		if err != nil {
			return fmt.Errorf("App still exists: %s", err)
		}
	}

	return nil
}

func testAccCheckKoyebAppAttributes(app *koyeb.App, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *app.Name != name {
			return fmt.Errorf("Bad name: %s", *app.Name)
		}

		return nil
	}
}

func testAccCheckKoyebAppExists(n string, app *koyeb.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.AppsApi.GetApp(context.Background(), rs.Primary.ID).Execute()

		if err != nil {
			return err
		}

		if *res.App.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*app = res.GetApp()

		return nil
	}
}

const testAccCheckKoyebAppConfig_basic = `
resource "koyeb_app" "foobar" {
	name       = "%s"
}`

func TestAppSchemaHasDeleteWhenEmpty(t *testing.T) {
	deleteWhenEmpty, ok := appSchema()["delete_when_empty"]
	if !ok {
		t.Fatal("expected delete_when_empty in the app schema")
	}
	if !deleteWhenEmpty.Optional {
		t.Error("expected delete_when_empty to be optional")
	}
	if !deleteWhenEmpty.Computed {
		t.Error("expected delete_when_empty to be computed: the API read-back is what lets a true->false change converge")
	}
}

func TestSetAppAttributeDeleteWhenEmpty(t *testing.T) {
	d := schema.TestResourceDataRaw(t, appSchema(), map[string]interface{}{})

	app := koyeb.App{
		Id:        toOpt("app-id"),
		Name:      toOpt("my-app"),
		LifeCycle: &koyeb.AppLifeCycle{DeleteWhenEmpty: toOpt(true)},
	}

	if err := setAppAttribute(d, app); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	if d.Get("delete_when_empty").(bool) != true {
		t.Errorf("expected delete_when_empty to be true, got %v", d.Get("delete_when_empty"))
	}
}

const testAccCheckKoyebAppConfig_delete_when_empty = `
resource "koyeb_app" "foobar" {
	name              = "%s"
	delete_when_empty = false
}`

func TestResourceKoyebAppUpdateSendsExplicitDeleteWhenEmptyFalse(t *testing.T) {
	var updateBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &updateBody)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"app":{"id":"d290f1ee-6c54-4b01-90e6-d7015f3f7b1f","name":"my-app"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	d := schema.TestResourceDataRaw(t, appSchema(), map[string]interface{}{
		"name":              "my-app",
		"delete_when_empty": false,
	})
	d.SetId("d290f1ee-6c54-4b01-90e6-d7015f3f7b1f")

	if diags := resourceKoyebAppUpdate(context.Background(), d, client); diags.HasError() {
		t.Fatalf("expected no error, got %v", diags)
	}

	lifeCycle, ok := updateBody["life_cycle"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected life_cycle to be sent on update, got body %v", updateBody)
	}
	if lifeCycle["delete_when_empty"] != false {
		t.Errorf("expected delete_when_empty false to be sent explicitly, got %v", lifeCycle["delete_when_empty"])
	}
}

func TestSetAppAttributeWithoutLifeCycle(t *testing.T) {
	d := schema.TestResourceDataRaw(t, appSchema(), map[string]interface{}{})

	app := koyeb.App{
		Id:   toOpt("app-id"),
		Name: toOpt("my-app"),
	}

	if err := setAppAttribute(d, app); err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	if d.Get("delete_when_empty").(bool) != false {
		t.Errorf("expected delete_when_empty to default to false without a life cycle, got %v", d.Get("delete_when_empty"))
	}
}

func TestResourceKoyebAppCreateSendsLifeCycle(t *testing.T) {
	var createBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &createBody)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"app":{"id":"d290f1ee-6c54-4b01-90e6-d7015f3f7b1f","name":"my-app"}}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	d := schema.TestResourceDataRaw(t, appSchema(), map[string]interface{}{
		"name":              "my-app",
		"delete_when_empty": true,
	})

	if diags := resourceKoyebAppCreate(context.Background(), d, client); diags.HasError() {
		t.Fatalf("expected no error, got %v", diags)
	}

	lifeCycle, ok := createBody["life_cycle"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected life_cycle to be sent on create, got body %v", createBody)
	}
	if lifeCycle["delete_when_empty"] != true {
		t.Errorf("expected delete_when_empty true to be sent, got %v", lifeCycle["delete_when_empty"])
	}
}

func TestDataSourceKoyebAppDeleteWhenEmptyIsReadOnly(t *testing.T) {
	deleteWhenEmpty := dataSourceKoyebApp().Schema["delete_when_empty"]

	if deleteWhenEmpty.Optional {
		t.Error("expected delete_when_empty to not be configurable on the data source")
	}
	if !deleteWhenEmpty.Computed {
		t.Error("expected delete_when_empty to be computed on the data source")
	}
}
