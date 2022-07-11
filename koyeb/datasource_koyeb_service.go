package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func dataSourceKoyebService() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKoyebServiceRead,
		Schema: map[string]*schema.Schema{
			"slug": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The service slug composed of the app and service name, for instance my-app/my-service",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the service",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the service",
				Computed:    true,
			},
			"app_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The app id the service is assigned",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization id owning the service",
				// Elem:        deploymentSchema(),
			},
			"active_deployment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service active deployment id",
				// Elem:        deploymentSchema(),
			},
			"latest_deployment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service latest deployment id",
				// Elem:        deploymentSchema(),
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version of the service",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of the service",
			},
			"messages": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The status messages of the service",
			},
			"paused_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"resumed_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"terminated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was created",
			},
		},
	}
}

func dataSourceKoyebServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(context.Background(), client)
	serviceMapper := mapper.Service()

	id, err := serviceMapper.ResolveID(d.Get("slug").(string))

	if err != nil {
		return diag.Errorf("Error retrieving service: %s", err)
	}

	d.SetId(id)

	return resourceKoyebServiceRead(ctx, d, meta)
}
