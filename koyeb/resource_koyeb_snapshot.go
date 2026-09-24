package koyeb

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func snapshotSchema() map[string]*schema.Schema {
	snapshot := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The snapshot ID",
		},
		"name": {
			Type:         schema.TypeString,
			Description:  "The snapshot name",
			Required:     true,
			ValidateFunc: validation.StringLenBetween(2, 64),
		},
		"parent_volume_id": {
			Type:        schema.TypeString,
			Description: "The ID of the volume the snapshot is taken from",
			ForceNew:    true,
			Required:    true,
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization ID owning the snapshot",
		},
		"size": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The size of the snapshot in GB",
		},
		"region": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The region where the snapshot is located",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the snapshot",
		},
		"type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The type of the snapshot",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the snapshot was last updated",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the snapshot was created",
		},
	}

	return snapshot
}

func resourceKoyebSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "Snapshot resource in the Koyeb Terraform provider.",

		CreateContext: resourceKoyebSnapshotCreate,
		ReadContext:   resourceKoyebSnapshotRead,
		UpdateContext: resourceKoyebSnapshotUpdate,
		DeleteContext: resourceKoyebSnapshotDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: snapshotSchema(),
	}
}

func setSnapshotAttribute(d *schema.ResourceData, snapshot koyeb.Snapshot) error {
	d.SetId(snapshot.GetId())
	d.Set("name", snapshot.GetName())
	d.Set("parent_volume_id", snapshot.GetParentVolumeId())
	d.Set("organization_id", snapshot.GetOrganizationId())
	d.Set("size", snapshot.GetSize())
	d.Set("region", snapshot.GetRegion())
	d.Set("status", snapshot.GetStatus())
	d.Set("type", snapshot.GetType())
	d.Set("updated_at", snapshot.GetUpdatedAt().UTC().String())
	d.Set("created_at", snapshot.GetCreatedAt().UTC().String())

	return nil
}

func resourceKoyebSnapshotCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.SnapshotsApi.CreateSnapshot(context.Background()).Body(koyeb.CreateSnapshotRequest{
		Name:           toOpt(d.Get("name").(string)),
		ParentVolumeId: toOpt(d.Get("parent_volume_id").(string)),
	}).Execute()

	if err != nil {
		return diag.Errorf("Error creating snapshot: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.Snapshot.Id)
	log.Printf("[INFO] Created snapshot name: %s", *res.Snapshot.Name)

	return resourceKoyebSnapshotRead(ctx, d, meta)
}

func resourceKoyebSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.SnapshotsApi.GetSnapshot(context.Background(), d.Id()).Execute()
	if err != nil {
		// If the snapshot is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving snapshot: %s (%v %v)", err, resp, res)
	}

	setSnapshotAttribute(d, *res.Snapshot)

	return nil
}

func resourceKoyebSnapshotUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.SnapshotsApi.UpdateSnapshot(context.Background(), d.Id()).Body(koyeb.UpdateSnapshotRequest{
		Name: toOpt(d.Get("name").(string)),
	}).Execute()

	if err != nil {
		return diag.Errorf("Error updating snapshot: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Updated snapshot name: %s", *res.Snapshot.Name)

	return resourceKoyebSnapshotRead(ctx, d, meta)
}

func resourceKoyebSnapshotDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.SnapshotsApi.DeleteSnapshot(context.Background(), d.Id()).Execute()

	if err != nil {
		return diag.Errorf("Error deleting snapshot: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
