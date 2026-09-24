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
	resource.AddTestSweepers("koyeb_snapshot", &resource.Sweeper{
		Name: "koyeb_snapshot",
		F:    testSweepSnapshot,
	})
	resource.AddTestSweepers("koyeb_volume", &resource.Sweeper{
		Name:         "koyeb_volume",
		F:            testSweepVolume,
		Dependencies: []string{"koyeb_service", "koyeb_snapshot"},
	})
}

func testSweepVolume(string) error {
	meta, err := sharedConfig()
	if err != nil {
		return err
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.PersistentVolumesApi.ListPersistentVolumes(context.Background()).Limit("100").Execute()
	if err != nil {
		return err
	}

	for _, v := range res.Volumes {
		if strings.HasPrefix(v.GetName(), testNamePrefix) {
			log.Printf("Destroying volume %s", v.GetName())

			if _, _, err := client.PersistentVolumesApi.DeletePersistentVolume(context.Background(), v.GetId()).Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func testSweepSnapshot(string) error {
	meta, err := sharedConfig()
	if err != nil {
		return err
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.SnapshotsApi.ListSnapshots(context.Background()).Limit("100").Execute()
	if err != nil {
		return err
	}

	for _, s := range res.Snapshots {
		if strings.HasPrefix(s.GetName(), testNamePrefix) {
			log.Printf("Destroying snapshot %s", s.GetName())

			if _, _, err := client.SnapshotsApi.DeleteSnapshot(context.Background(), s.GetId()).Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestSnapshotSchemaContract(t *testing.T) {
	s := snapshotSchema()

	name := s["name"]
	if !name.Required {
		t.Error("expected name to be required")
	}
	if name.ForceNew {
		t.Error("expected name to be updatable")
	}

	parentVolumeID := s["parent_volume_id"]
	if !parentVolumeID.Required {
		t.Error("expected parent_volume_id to be required")
	}
	if !parentVolumeID.ForceNew {
		t.Error("expected parent_volume_id to force new resources")
	}

	for _, key := range []string{"id", "organization_id", "size", "region", "status", "type", "updated_at", "created_at"} {
		field, ok := s[key]
		if !ok {
			t.Errorf("expected %q in the snapshot schema", key)
			continue
		}
		if !field.Computed {
			t.Errorf("expected %q to be computed", key)
		}
	}
}

func TestSnapshotDataSourceSchemaContract(t *testing.T) {
	s := snapshotDataSourceSchema()

	name := s["name"]
	if !name.Required {
		t.Error("expected name to be required to look up a snapshot")
	}

	for _, key := range []string{"id", "parent_volume_id", "organization_id", "size", "region", "status", "type", "updated_at", "created_at"} {
		field, ok := s[key]
		if !ok {
			t.Errorf("expected %q in the snapshot datasource schema", key)
			continue
		}
		if !field.Computed {
			t.Errorf("expected %q to be computed", key)
		}
	}
}

func TestAccKoyebSnapshot_Basic(t *testing.T) {
	var snapshot koyeb.Snapshot
	volumeName := randomTestName()
	appName := randomTestName()
	snapshotName := randomTestName()
	renamedSnapshotName := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebSnapshotConfig_basic, volumeName, appName, snapshotName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebSnapshotExists("koyeb_snapshot.foobar", &snapshot),
					resource.TestCheckResourceAttr("koyeb_snapshot.foobar", "name", snapshotName),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "parent_volume_id"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "size"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "region"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "type"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_snapshot.foobar", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebSnapshotConfig_basic, volumeName, appName, renamedSnapshotName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebSnapshotExists("koyeb_snapshot.foobar", &snapshot),
					resource.TestCheckResourceAttr("koyeb_snapshot.foobar", "name", renamedSnapshotName),
				),
			},
			{
				Config:                  fmt.Sprintf(testAccCheckKoyebSnapshotConfig_basic, volumeName, appName, renamedSnapshotName),
				ImportState:             true,
				ResourceName:            "koyeb_snapshot.foobar",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
			},
		},
	})
}

func testAccCheckKoyebSnapshotDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_snapshot" {
			continue
		}

		err := waitForResourceStatus(client.SnapshotsApi.GetSnapshot(context.Background(), rs.Primary.ID).Execute, "Snapshot", targetStatus, 1, false)
		if err != nil {
			return fmt.Errorf("Snapshot still exists: %s", err)
		}
	}

	return nil
}

func testAccCheckKoyebSnapshotExists(n string, snapshot *koyeb.Snapshot) resource.TestCheckFunc {
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

const testAccCheckKoyebSnapshotConfig_basic = `
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

# Snapshots can only be taken from volumes mounted to a service.
resource "koyeb_snapshot" "foobar" {
	name             = "%s"
	parent_volume_id = koyeb_volume.foobar.id

	depends_on = [
	  koyeb_service.bar
	]
}`
