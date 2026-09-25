package koyeb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

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

			// Services are swept first, but the detach is asynchronous.
			if err := deleteVolumeWhenDetached(client, v.GetId(), 10*time.Second); err != nil {
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
	// Services with mounted volumes schedule on a limited host pool; run
	// on one Terraform matrix version to avoid six concurrent volume
	// deployments racing for capacity.
	if v := os.Getenv("TF_ACC_MATRIX"); v != "" && !strings.HasPrefix(v, "1.1") {
		t.Skipf("skipping to limit concurrent volume deployments (TF %s)", v)
	}

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
				// Snapshots require a volume that is mounted, and the mount
				// only happens once the service is deployed: wait for the
				// service to be healthy before adding the snapshot.
				Config: fmt.Sprintf(testAccCheckKoyebSnapshotConfig_mounted, volumeName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceHealthy("koyeb_service.bar"),
				),
			},
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

// testAccCheckKoyebServiceHealthy polls the service until it reports
// HEALTHY, so a dependent resource (like a volume snapshot) is only created
// after the volume is actually mounted.
func testAccCheckKoyebServiceHealthy(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		var lastStatus koyeb.ServiceStatus
		for i := 0; i < 45; i++ {
			res, _, err := client.ServicesApi.GetService(context.Background(), rs.Primary.ID).Execute()
			if err != nil {
				return err
			}
			service := res.GetService()
			lastStatus = service.GetStatus()
			if service.GetStatus() == koyeb.SERVICESTATUS_HEALTHY {
				return nil
			}
			time.Sleep(10 * time.Second)
		}

		return fmt.Errorf("service %s did not become healthy in time, last status %s", n, lastStatus)
	}
}

func testAccCheckKoyebSnapshotDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"SNAPSHOT_STATUS_DELETED", "SNAPSHOT_STATUS_DELETING"}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_snapshot" {
			continue
		}

		err := waitForStatus(context.Background(), goneWait("Snapshot", targetStatus, time.Minute), snapshotStatusPoller(context.Background(), client, rs.Primary.ID))
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

const testAccCheckKoyebSnapshotConfig_mounted = `
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
		ports {
		  port     = 3000
		  protocol = "http"
		}
		scalings {
		  min = 1
		  max = 1
		}
		routes {
		  port = 3000
		  path = "/"
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
`

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

func TestDeleteSnapshotWhenUploadedRetriesWhileUploading(t *testing.T) {
	const snapshotID = "d290f1ee-6c54-4b01-90e6-d7015f3f7b1f"
	deletes := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "DELETE" {
			_, _ = w.Write([]byte(`{"snapshot":{"id":"` + snapshotID + `","status":"SNAPSHOT_STATUS_CREATING"}}`))
			return
		}
		deletes++
		if deletes < 2 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":400,"code":"failed_precondition","message":"cannot delete a snapshot being uploaded"}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	client := koyeb.NewAPIClient(cfg)

	if err := deleteSnapshotWhenUploaded(client, snapshotID, 10*time.Millisecond); err != nil {
		t.Fatalf("expected the delete to eventually succeed, got %s", err)
	}
	if deletes != 2 {
		t.Errorf("expected 2 delete attempts, got %d", deletes)
	}
}
