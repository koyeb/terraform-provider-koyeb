package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func snapshotDataSourceSchema() map[string]*schema.Schema {
	snapshot := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The snapshot ID",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The snapshot name",
		},
		"parent_volume_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The ID of the volume the snapshot is taken from",
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

func dataSourceKoyebSnapshot() *schema.Resource {
	return &schema.Resource{
		Description: "Snapshot data source in the Koyeb Terraform provider.",
		ReadContext: dataSourceKoyebSnapshotRead,
		Schema:      snapshotDataSourceSchema(),
	}
}

func dataSourceKoyebSnapshotRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(context.Background(), client)
	snapshotMapper := mapper.Snapshot()

	id, err := snapshotMapper.ResolveID(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("Error retrieving snapshot: %s", err)
	}

	d.SetId(id)

	return resourceKoyebSnapshotRead(ctx, d, meta)
}
