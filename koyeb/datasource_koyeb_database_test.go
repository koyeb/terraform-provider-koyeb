package koyeb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccDataSourceKoyebDatabase_Basic(t *testing.T) {
	// See TestAccKoyebDatabase_Basic: one matrix version only.
	if v := os.Getenv("TF_ACC_MATRIX"); v != "" && !strings.HasPrefix(v, "1.1") {
		t.Skipf("skipping to stay within the organization's free instance quota (TF %s)", v)
	}

	var service koyeb.Service
	databaseName := randomTestName()

	resourceConfig := fmt.Sprintf(`
resource "koyeb_database" "foobar" {
	name = "%s"
}
`, databaseName)

	dataSourceConfig := `
data "koyeb_database" "bar" {
	name = koyeb_database.foobar.name
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if exhausted, err := freeInstanceQuotaExhausted(testAccProvider.Meta().(*koyeb.APIClient)); err == nil && exhausted {
				t.Skip("skipping: the organization's free instance quota is exhausted")
			}
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
			},
			{
				Config: resourceConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceKoyebDatabaseExists("data.koyeb_database.bar", &service),
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

func testAccCheckDataSourceKoyebDatabaseExists(n string, service *koyeb.Service) resource.TestCheckFunc {
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
