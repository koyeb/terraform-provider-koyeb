package koyeb

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func servicePoolSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The service pool ID",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The service pool name",
		},
		"size": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "The number of prewarmed services kept ready in the pool",
		},
		"definition": {
			Type:        schema.TypeList,
			MinItems:    1,
			MaxItems:    1,
			Required:    true,
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
	}
}

func resourceKoyebServicePool() *schema.Resource {
	return &schema.Resource{
		Description: "Service pool resource in the Koyeb Terraform provider. " +
			"Service pools keep a set of prewarmed services ready to be claimed instantly. " +
			"The Koyeb API does not expose update or delete endpoints for service pools yet, " +
			"so this resource cannot be modified or destroyed in place.",

		CreateContext: resourceKoyebServicePoolCreate,
		ReadContext:   resourceKoyebServicePoolRead,
		UpdateContext: resourceKoyebServicePoolUpdate,
		DeleteContext: resourceKoyebServicePoolDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: servicePoolSchema(),
	}
}

func setServicePoolAttribute(d *schema.ResourceData, pool koyeb.ServicePool) error {
	d.SetId(pool.GetId())
	d.Set("name", pool.GetName())
	d.Set("size", int(pool.GetSize()))
	if definition, ok := pool.GetDefinitionOk(); ok && definition != nil {
		d.Set("definition", flattenDeploymentDefinition(definition))
	}
	d.Set("organization_id", pool.GetOrganizationId())
	d.Set("workspace_id", pool.GetWorkspaceId())
	d.Set("ready_count", int(pool.GetReadyCount()))
	d.Set("status", pool.GetStatus())
	d.Set("messages", strings.Join(pool.GetMessages(), " "))
	d.Set("generation", pool.GetGeneration())
	d.Set("customer_id", pool.GetCustomerId())
	d.Set("updated_at", pool.GetUpdatedAt().UTC().String())
	d.Set("created_at", pool.GetCreatedAt().UTC().String())

	return nil
}

func resourceKoyebServicePoolCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	definition := expandDeploymentDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{}))

	res, resp, err := client.ServicePoolsApi.CreateServicePool(context.Background()).ServicePool(koyeb.CreateServicePool{
		Name:       toOpt(d.Get("name").(string)),
		Size:       toOpt(int64(d.Get("size").(int))),
		Definition: definition,
	}).Execute()
	if err != nil {
		return diag.Errorf("Error creating service pool: %s (%v %v)", err, resp, res)
	}

	pool := res.GetServicePool()
	d.SetId(pool.GetId())
	log.Printf("[INFO] Created service pool name: %s", pool.GetName())

	return resourceKoyebServicePoolRead(ctx, d, meta)
}

func resourceKoyebServicePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.ServicePoolsApi.GetServicePool(context.Background(), d.Id()).Execute()
	if err != nil {
		// If the service pool is somehow already destroyed, mark as
		// successfully gone
		if resp != nil && resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving service pool: %s (%v %v)", err, resp, res)
	}

	setServicePoolAttribute(d, res.GetServicePool())

	return nil
}

// NOTE: The Koyeb API does not expose an update endpoint for service pools
// yet (PUT /v1/service_pools/{id} is still under development upstream).
// Terraform requires UpdateContext to be set, so fail explicitly instead of
// pretending the update succeeded.
func resourceKoyebServicePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf(
		"Updating service pools is not yet supported by the Koyeb API (no update endpoint is available yet), pool: %s. "+
			"Apply your change out-of-band with the Koyeb console or API and re-import the resource, "+
			"or wait for update support to land.",
		d.Id(),
	)
}

// NOTE: The Koyeb API does not expose a delete endpoint for service pools
// yet (DELETE /v1/service_pools/{id} is still under development upstream).
// Terraform requires DeleteContext to be set, so fail explicitly instead of
// faking a successful destroy: the pool is left untouched in the Koyeb
// organization and remains in the Terraform state.
func resourceKoyebServicePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf(
		"Deleting service pools is not yet supported by the Koyeb API (no delete endpoint is available yet), "+
			"pool: %s was not deleted. "+
			"If you no longer want to manage it with Terraform, remove it from the state with `terraform state rm`.",
		d.Id(),
	)
}
