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
	resource.AddTestSweepers("koyeb_service", &resource.Sweeper{
		Name:         "koyeb_service",
		F:            testSweepService,
		Dependencies: []string{"koyeb_app"},
	})

}

func testSweepService(string) error {
	meta, err := sharedConfig()
	if err != nil {
		return err
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.ServicesApi.ListServices(context.Background()).Limit("100").Execute()
	if err != nil {
		return err
	}

	for _, a := range *res.Services {
		if strings.HasPrefix(a.GetName(), testNamePrefix) {
			log.Printf("Destroying service %s", *a.Name)

			if _, _, err := client.ServicesApi.DeleteService(context.Background(), a.GetId()).Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccKoyebService_Basic(t *testing.T) {
	var service koyeb.Service
	appName := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_basic_docker, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "docker"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "created_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "app_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "version"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "messages"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "paused_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "resumed_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "terminated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "latest_deployment"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_basic_git, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "git"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "created_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "app_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "version"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "messages"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "paused_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "resumed_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "terminated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "latest_deployment"),
				),
			},
		},
	})
}

func testAccCheckKoyebServiceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"DELETED", "DELETING"}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_service" {
			continue
		}

		err := waitForResourceStatus(client.ServicesApi.GetService(context.Background(), rs.Primary.ID).Execute, "Service", targetStatus, 1, false)
		if err != nil {
			return fmt.Errorf("Service still exists: %s ", err)
		}

	}

	return nil
}

func testAccCheckKoyebServiceExists(n string, service *koyeb.Service) resource.TestCheckFunc {
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

		a := res.GetService()
		*service = a

		return nil
	}
}

const testAccCheckKoyebServiceConfig_basic_docker = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "docker"
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
		regions = ["fra"]
		docker {
		  image = "koyeb/demo"
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`

const testAccCheckKoyebServiceConfig_basic_git = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "git"
		instance_types {
		  type = "micro"
		}
		ports {
		  port     = 8080
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
		  port = 8080
		}
		regions = ["fra"]
		git {
		  repository = "github.com/koyeb/example-flask"
		  branch = "main"
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`
