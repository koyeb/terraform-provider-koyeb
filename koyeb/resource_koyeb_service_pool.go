package koyeb

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// Readiness budget for created and updated pools; a variable so tests
// can shorten it. ERROR and DELETING are terminal: neither pool ever
// becomes READY.
var (
	servicePoolReadinessTimeout = 5 * time.Minute
	servicePoolTerminalStatuses = []string{"ERROR", "DELETING"}
)

func waitForPoolReady(ctx context.Context, client *koyeb.APIClient, poolID string) error {
	return waitForStatus(ctx, statusWait{
		name:      "ServicePool",
		targets:   []string{"READY"},
		terminals: servicePoolTerminalStatuses,
		timeout:   servicePoolReadinessTimeout,
		interval:  waitRetryInterval,
	}, servicePoolStatusPoller(ctx, client, poolID))
}

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
			Elem:        deploymentDefinitionSchema(),
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
			"Size and definition changes are applied in place; " +
			"renaming a pool is not supported by the Koyeb API. " +
			"Create and update wait for the pool to become READY before completing. " +
			"Sandbox pools need definition type = \"SANDBOX\" set explicitly (the provider defaults to WEB, unlike the SDKs) " +
			"and, on the koyeb/sandbox image, a SANDBOX_SECRET env var supplied via definition env.",

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

	res, resp, err := client.ServicePoolsApi.CreateServicePool(ctx).ServicePool(koyeb.CreateServicePool{
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

	// Pools prewarm asynchronously: apply should only report success once
	// the pool, not just the API call, is READY.
	if err := waitForPoolReady(ctx, client, d.Id()); err != nil {
		return diag.Errorf("Error waiting for service pool to be ready: %s", err)
	}

	return resourceKoyebServicePoolRead(ctx, d, meta)
}

func resourceKoyebServicePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.ServicePoolsApi.GetServicePool(ctx, d.Id()).Execute()
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

// NOTE: The update endpoint only accepts size and definition: the API cannot
// rename a pool, so fail explicitly on a name change instead of silently
// ignoring it. The shared definition block keeps ForceNew: true on its
// nested `name` (deploymentDefinitionSchema in resource_koyeb_service.go),
// so renaming the definition plans a replacement; that apply fails at the
// destroy step with the explicit delete error while deletion stays
// unsupported upstream.
func resourceKoyebServicePoolUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("name") {
		return diag.Errorf(
			"Renaming service pools is not supported by the Koyeb API, pool: %s. "+
				"Revert the name change in your configuration or recreate the pool once deletion support lands.",
			d.Id(),
		)
	}

	client := meta.(*koyeb.APIClient)

	definition := expandDeploymentDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{}))

	res, resp, err := client.ServicePoolsApi.UpdateServicePool(ctx, d.Id()).
		ServicePool(koyeb.UpdateServicePool{
			Size:       toOpt(int64(d.Get("size").(int))),
			Definition: definition,
		}).Execute()
	if err != nil {
		return diag.Errorf("Error updating service pool: %s (%v %v)", err, resp, res)
	}

	pool := res.GetServicePool()
	log.Printf("[INFO] Updated service pool name: %s", pool.GetName())

	if err := waitForPoolReady(ctx, client, d.Id()); err != nil {
		return diag.Errorf("Error waiting for service pool to be ready: %s", err)
	}

	return resourceKoyebServicePoolRead(ctx, d, meta)
}

// NOTE: Upstream deletion is asynchronous: the API marks the pool
// is_deleting and replies immediately, so Terraform forgets the resource
// while reclamation happens server-side.
func resourceKoyebServicePoolDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.ServicePoolsApi.DeleteServicePool(ctx, d.Id()).Execute()
	if err != nil {
		return diag.Errorf("Error deleting service pool: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
