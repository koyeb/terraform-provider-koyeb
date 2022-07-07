package koyeb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

	for _, a := range *res.Apps {
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
		},
	})
}

func testAccCheckKoyebAppDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_app" {
			continue
		}

		_, _, err := client.AppsApi.GetApp(context.Background(), rs.Primary.ID).Execute()
		if err == nil {
			return fmt.Errorf("App still exists: %s ", err)
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

		a := res.GetApp()
		*app = a

		return nil
	}
}

const testAccCheckKoyebAppConfig_basic = `
resource "koyeb_app" "foobar" {
	name       = "%s"
}`
