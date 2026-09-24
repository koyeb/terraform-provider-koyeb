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

func TestAccDataSourceKoyebSnapshot_Basic(t *testing.T) {
	// See TestAccKoyebSnapshot_Basic: one matrix version only.
	if v := os.Getenv("TF_ACC_MATRIX"); v != "" && !strings.HasPrefix(v, "1.1") {
		t.Skipf("skipping to limit concurrent volume deployments (TF %s)", v)
	}

	var snapshot koyeb.Snapshot
	volumeName := randomTestName()
	appName := randomTestName()
	snapshotName := randomTestName()

	resourceConfigTemplate := fmt.Sprintf(`
resource "koyeb_volume" "foobar" {
	name     = "%s"
	max_size = 1
	region   = "was"
}

resource "koyeb_app" "app" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = koyeb_app.app.name
	definition {
		name = "service"
		instance_types {
		  type = "micro"
		}
		scalings {
		  min = 1
		  max = 1
		}
		volumes {
		  id   = koyeb_volume.foobar.id
		  path = "/data"
		}
		regions = ["was"]
		docker {
		  image = "koyeb/demo"
		}
	}

	depends_on = [
	  koyeb_app.app
	]
}
`, volumeName, appName)

	snapshotConfig := fmt.Sprintf(`
resource "koyeb_snapshot" "foobar" {
	name             = "%s"
	parent_volume_id = koyeb_volume.foobar.id

	depends_on = [
	  koyeb_service.bar
	]
}
`, snapshotName)

	dataSourceConfig := `
data "koyeb_snapshot" "bar" {
	name = koyeb_snapshot.foobar.name
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				// Wait for the service to mount the volume before adding
				// the snapshot and the data source.
				Config: resourceConfigTemplate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceHealthy("koyeb_service.bar"),
				),
			},
			{
				Config: resourceConfigTemplate + snapshotConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceKoyebSnapshotExists("data.koyeb_snapshot.bar", &snapshot),
					resource.TestCheckResourceAttr("data.koyeb_snapshot.bar", "name", snapshotName),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "id"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "parent_volume_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "size"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "region"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "status"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "type"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_snapshot.bar", "created_at"),
				),
			},
		},
	})
}

func testAccCheckDataSourceKoyebSnapshotExists(n string, snapshot *koyeb.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.SnapshotsApi.GetSnapshot(context.Background(), rs.Primary.ID).Execute()

		if err != nil {
			return err
		}

		if *res.Snapshot.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*snapshot = *res.Snapshot

		return nil
	}
}
