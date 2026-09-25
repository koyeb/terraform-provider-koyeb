package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestVolumeDataSourceSchemaContract(t *testing.T) {
	s := volumeDataSourceSchema()

	name := s["name"]
	if !name.Required {
		t.Error("expected name to be required to look up a volume")
	}

	for _, key := range []string{"id", "organization_id", "snapshot_id", "service_id", "region", "read_only", "max_size", "cur_size", "status", "backing_store", "volume_type", "updated_at", "created_at"} {
		field, ok := s[key]
		if !ok {
			t.Errorf("expected %q in the volume datasource schema", key)
			continue
		}
		if !field.Computed {
			t.Errorf("expected %q to be computed", key)
		}
	}
}

func TestAccDataSourceKoyebVolume_Basic(t *testing.T) {
	var volume koyeb.PersistentVolume
	volumeName := randomTestName()

	resourceConfig := fmt.Sprintf(`
resource "koyeb_volume" "foobar" {
	name     = "%s"
	max_size = 1
	region   = "was"
}
`, volumeName)

	dataSourceConfig := `
data "koyeb_volume" "bar" {
	name = koyeb_volume.foobar.name
}`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
			},
			{
				Config: resourceConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceKoyebVolumeExists("data.koyeb_volume.bar", &volume),
					resource.TestCheckResourceAttr("data.koyeb_volume.bar", "name", volumeName),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "id"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "region"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "max_size"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "cur_size"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "status"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "backing_store"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "volume_type"),
					resource.TestCheckResourceAttr("data.koyeb_volume.bar", "read_only", "false"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_volume.bar", "created_at"),
				),
			},
		},
	})
}

func testAccCheckDataSourceKoyebVolumeExists(n string, volume *koyeb.PersistentVolume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.PersistentVolumesApi.GetPersistentVolume(context.Background(), rs.Primary.ID).Execute()

		if err != nil {
			return err
		}

		if *res.Volume.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*volume = *res.Volume

		return nil
	}
}
