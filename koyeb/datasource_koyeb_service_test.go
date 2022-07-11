package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccDataSourceKoyebService_Basic(t *testing.T) {
	var service koyeb.Service
	appName := randomTestName()

	resourceConfig := fmt.Sprintf(`
resource "koyeb_app" "foo" {
  name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "main"
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
		env {
		  key   = "FOO"
		  value = "BAR"
		}
		routes {
		  path = "/"
		  port = 3000
		}
		regions = ["par"]
		docker {
		  image = "koyeb/demo"
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`, appName, appName)

	dataSourceConfig := `
data "koyeb_service" "foobar" {
  slug = "${koyeb_app.foo.name}/${koyeb_service.bar.name}"
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
					testAccCheckDataSourceKoyebServiceExists("data.koyeb_service.foobar", &service),
					resource.TestCheckResourceAttr("data.koyeb_service.foobar", "name", "main"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "id"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "organization_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "updated_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "created_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "app_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "version"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "status"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "messages"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "paused_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "resumed_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "terminated_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_service.foobar", "latest_deployment"),
				),
			},
		},
	})
}

func testAccCheckDataSourceKoyebServiceExists(n string, service *koyeb.Service) resource.TestCheckFunc {
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

		*service = res.GetService()

		return nil
	}
}
