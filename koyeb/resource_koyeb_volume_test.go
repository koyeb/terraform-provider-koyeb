package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccKoyebVolume_Basic(t *testing.T) {
	var volume koyeb.PersistentVolume
	volumeName := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebVolumeConfig_basic, volumeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebVolumeExists("koyeb_volume.foobar", &volume),
					testAccCheckKoyebVolumeAttributes(&volume, volumeName),
					resource.TestCheckResourceAttr(
						"koyeb_volume.foobar", "name", volumeName),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "volume_type"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "name"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "region"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "read_only"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "max_size"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "cur_size"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "backing_store"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_volume.foobar", "created_at"),
				),
			},
		},
	})
}

func testAccCheckKoyebVolumeDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"DELETED", "DELETING"}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_volume" {
			continue
		}

		err := waitForResourceStatus(client.PersistentVolumesApi.GetPersistentVolume(context.Background(), rs.Primary.ID).Execute, "Volume", targetStatus, 1, false)
		if err == nil {
			return fmt.Errorf("Volume still exists: %s ", err)
		}
	}

	return nil
}

func testAccCheckKoyebVolumeAttributes(volume *koyeb.PersistentVolume, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *volume.Name != name {
			return fmt.Errorf("Bad name: %s", *volume.Name)
		}

		return nil
	}
}

func testAccCheckKoyebVolumeExists(n string, volume *koyeb.PersistentVolume) resource.TestCheckFunc {
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

		*volume = res.GetVolume()

		return nil
	}
}

const testAccCheckKoyebVolumeConfig_basic = `
resource "koyeb_volume" "foobar" {
	name       = "%s"
	max_size   = 10
	region     = "was"
}`
