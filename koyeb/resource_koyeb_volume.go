package koyeb

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func volumeSchema() map[string]*schema.Schema {
	volume := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The volume ID",
		},
		"volume_type": {
			Type:        schema.TypeString,
			Description: "The volume type",
			ForceNew:    true,
			Optional:    true,
			Default:     "PERSISTENT_VOLUME_BACKING_STORE_LOCAL_BLK",
			ValidateFunc: validation.StringInSlice([]string{
				"PERSISTENT_VOLUME_BACKING_STORE_INVALID",
				"PERSISTENT_VOLUME_BACKING_STORE_LOCAL_BLK",
			}, false),
		},
		"name": {
			Type:         schema.TypeString,
			Description:  "The volume name",
			Required:     true,
			ValidateFunc: validation.StringLenBetween(2, 64),
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization ID owning the volume",
		},
		"snapshot_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The snapshot ID the volume was created from",
		},
		"service_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The service ID the volume is attached to",
		},
		"region": {
			Type:        schema.TypeString,
			Description: "The region where the volume is located",
			ForceNew:    true,
			Required:    true,
		},
		"read_only": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "If set to true, the volume will be mounted in read-only",
		},
		"max_size": {
			Type:        schema.TypeInt,
			Description: "The maximum size of the volume in GB",
			ForceNew:    true,
			Required:    true,
		},
		"cur_size": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The current size of the volume in GB",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the volume",
		},
		"backing_store": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The backing store of the volume",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the volume was last updated",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the volume was created",
		},
	}

	return volume
}

func resourceKoyebVolume() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Volume resource in the Koyeb Terraform provider.",

		CreateContext: resourceKoyebVolumeCreate,
		ReadContext:   resourceKoyebVolumeRead,
		UpdateContext: resourceKoyebVolumeUpdate,
		DeleteContext: resourceKoyebVolumeDelete,

		Schema: volumeSchema(),
	}
}

func setVolumeAttribute(d *schema.ResourceData, volume koyeb.PersistentVolume) error {
	d.SetId(volume.GetId())
	d.Set("name", volume.GetName())
	d.Set("max_size", volume.GetMaxSize())
	d.Set("region", volume.GetRegion())
	d.Set("snapshot_id", volume.GetSnapshotId())
	d.Set("service_id", volume.GetServiceId())
	d.Set("read_only", volume.GetReadOnly())
	d.Set("cur_size", volume.GetCurSize())
	d.Set("status", volume.GetStatus())
	d.Set("backing_store", volume.GetBackingStore())
	d.Set("organization_id", volume.GetOrganizationId())
	d.Set("created_at", volume.GetCreatedAt().UTC().String())
	d.Set("updated_at", volume.GetUpdatedAt().UTC().String())

	return nil
}

func resourceKoyebVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.PersistentVolumesApi.CreatePersistentVolume(context.Background()).Body(koyeb.CreatePersistentVolumeRequest{
		Name:       toOpt(d.Get("name").(string)),
		VolumeType: toOpt(koyeb.PersistentVolumeBackingStore(d.Get("volume_type").(string))),
		MaxSize:    toOpt(int64(d.Get("max_size").(int))),
		Region:     toOpt(d.Get("region").(string)),
	}).Execute()
	if err != nil {
		return diag.Errorf("Error creating volume: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Created volume name: %s", *res.Volume.Name)

	setVolumeAttribute(d, *res.Volume)

	return resourceKoyebVolumeRead(ctx, d, meta)
}

func resourceKoyebVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.PersistentVolumesApi.GetPersistentVolume(context.Background(), d.Id()).Execute()
	if err != nil {
		// If the volume is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving volume: %s (%v %v)", err, resp, res)
	}

	setVolumeAttribute(d, *res.Volume)

	return nil
}

func resourceKoyebVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.PersistentVolumesApi.UpdatePersistentVolume(context.Background(), d.Id()).Body(koyeb.UpdatePersistentVolumeRequest{
		Name:    toOpt(d.Get("name").(string)),
		MaxSize: toOpt(int64(d.Get("max_size").(int))),
	}).Execute()

	if err != nil {
		return diag.Errorf("Error updating volume: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Updated volume name: %s", *res.Volume.Name)
	return resourceKoyebVolumeRead(ctx, d, meta)
}

func resourceKoyebVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.PersistentVolumesApi.DeletePersistentVolume(context.Background(), d.Id()).Execute()

	if err != nil {
		return diag.Errorf("Error deleting volume: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
