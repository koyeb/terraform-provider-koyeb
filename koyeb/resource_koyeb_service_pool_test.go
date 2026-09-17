package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// NOTE: No sweeper is registered for service pools and no CheckDestroy is
// asserted: the Koyeb API does not expose DELETE /v1/service_pools/{id} yet,
// so pools created by these tests cannot be destroyed, and the implicit
// destroy step of the acceptance test harness fails with the explicit error
// returned by resourceKoyebServicePoolDelete. Run these tests only if you
// accept leaving the created pool in the organization until deletion
// support lands in the Koyeb API.

func TestAccKoyebServicePool_Basic(t *testing.T) {
	var pool koyeb.ServicePool
	poolName := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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
