package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func dataSourceKoyebServicePool() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKoyebServicePoolRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The service pool name",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service pool ID",
			},
			"size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of prewarmed services kept ready in the pool",
			},
			"definition": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The deployment definition of the services provisioned by the pool",
				Elem:        deploymentDefinitionSchena(),
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization ID owning the service pool",
			},
			"workspace_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The workspace ID owning the service pool",
			},
			"ready_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of services currently ready in the pool",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service pool status",
			},
			"messages": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service pool status messages",
			},
			"generation": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current generation of the service pool",
			},
			"customer_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The customer ID owning the service pool",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service pool was last updated",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service pool was created",
			},
		},
	}
}

func dataSourceKoyebServicePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	name := d.Get("name").(string)

	res, resp, err := client.ServicePoolsApi.ListServicePools(context.Background()).Name(name).Execute()
	if err != nil {
		return diag.Errorf("Error retrieving service pool: %s (%v %v)", err, resp, res)
	}

	for _, pool := range res.GetServicePools() {
		if pool.GetName() == name {
			d.SetId(pool.GetId())
			return resourceKoyebServicePoolRead(ctx, d, meta)
		}
	}

	return diag.Errorf("No service pool found with name %s", name)
}
