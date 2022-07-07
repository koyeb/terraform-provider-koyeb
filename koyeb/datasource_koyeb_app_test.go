package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccDataSourceKoyebApp_Basic(t *testing.T) {
	var app koyeb.App
	appName := randomTestName()

	resourceConfig := fmt.Sprintf(`
resource "koyeb_app" "foo" {
  name = "%s"
}
`, appName)

	dataSourceConfig := `
data "koyeb_app" "bar" {
  name = koyeb_app.foo.name
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
			},
			{
				Config: resourceConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceKoyebAppExists("data.koyeb_app.bar", &app),
					testAccCheckDataSourceKoyebAppAttributes(&app, appName),
					resource.TestCheckResourceAttr(
						"data.koyeb_app.bar", "name", appName),
				),
			},
		},
	})
}

func testAccCheckDataSourceKoyebAppAttributes(app *koyeb.App, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *app.Name != name {
			return fmt.Errorf("Bad name: %s", *app.Name)
		}

		return nil
	}
}

func testAccCheckDataSourceKoyebAppExists(n string, app *koyeb.App) resource.TestCheckFunc {
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
