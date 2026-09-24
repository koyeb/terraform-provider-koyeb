package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func volumeDataSourceSchema() map[string]*schema.Schema {
	volume := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The volume ID",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The volume name",
		},
		"volume_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The volume type",
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
			Computed:    true,
			Description: "The region where the volume is located",
		},
		"read_only": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "If set to true, the volume is mounted in read-only",
		},
		"max_size": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The maximum size of the volume in GB",
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

func dataSourceKoyebVolume() *schema.Resource {
	return &schema.Resource{
		Description: "Volume data source in the Koyeb Terraform provider.",
		ReadContext: dataSourceKoyebVolumeRead,
		Schema:      volumeDataSourceSchema(),
	}
}

func dataSourceKoyebVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(context.Background(), client)
	volumeMapper := mapper.Volume()

	id, err := volumeMapper.ResolveID(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("Error retrieving volume: %s", err)
	}

	d.SetId(id)

	return resourceKoyebVolumeRead(ctx, d, meta)
}
