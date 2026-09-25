package koyeb

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func serviceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The service ID",
		},
		"name": {
			Type:        schema.TypeString,
			Description: "The service name",
			Computed:    true,
		},
		"app_name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "The app name the service is assigned to",
			ValidateFunc: validation.StringLenBetween(3, 23),
		},
		"app_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The app id the service is assigned to",
		},
		"definition": {
			Type:        schema.TypeList,
			MinItems:    1,
			MaxItems:    1,
			Required:    true,
			Description: "The service deployment definition",
			Elem:        deploymentDefinitionSchema(),
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization ID owning the service",
		},
		"active_deployment": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The service active deployment ID",
		},
		"latest_deployment": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The service latest deployment ID",
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
	}
}

var configFilePermissionsRegex = regexp.MustCompile(`^0[0-7]{3}$`)

var serviceReadinessTimeout = 10 * time.Minute

// serviceReadyStatuses mirrors the Python SDK's classify_service_status:
// HEALTHY and DEGRADED are usable, STARTING and RESUMING keep polling, and
// every other known state is terminal (unknown statuses poll to the
// timeout).

var (
	serviceReadyStatuses    = []string{"HEALTHY", "DEGRADED"}
	serviceTerminalStatuses = []string{"UNHEALTHY", "DELETING", "DELETED", "PAUSING", "PAUSED"}
)

func waitForServiceReady(ctx context.Context, client *koyeb.APIClient, serviceID string, timeout time.Duration) error {
	return waitForStatus(ctx, statusWait{
		name:      "Service",
		targets:   serviceReadyStatuses,
		terminals: serviceTerminalStatuses,
		timeout:   timeout,
		interval:  waitRetryInterval,
	}, serviceStatusPoller(ctx, client, serviceID))
}

func resourceKoyebService() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Service resource in the Koyeb Terraform provider. " +
			"Create and update wait for the service to become HEALTHY or DEGRADED before completing.",

		CreateContext: resourceKoyebServiceCreate,
		ReadContext:   resourceKoyebServiceRead,
		UpdateContext: resourceKoyebServiceUpdate,
		DeleteContext: resourceKoyebServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: serviceSchema(),
	}
}

func setServiceAttribute(
	d *schema.ResourceData,
	service *koyeb.Service,
	// latestDeployment *koyeb.Deployment,
) error {
	d.SetId(service.GetId())
	d.Set("name", service.GetName())
	d.Set("app_id", service.GetAppId())
	// d.Set("definition", flattenDeploymentDefinition(toOpt(latestDeployment.GetDefinition())))
	d.Set("organization_id", service.GetOrganizationId())
	d.Set("active_deployment", service.GetActiveDeploymentId())
	d.Set("latest_deployment", service.GetLatestDeploymentId())
	d.Set("version", service.GetVersion())
	d.Set("status", service.GetStatus())
	d.Set("messages", strings.Join(service.GetMessages(), " "))
	d.Set("paused_at", service.GetPausedAt().UTC().String())
	d.Set("resumed_at", service.GetResumedAt().UTC().String())
	d.Set("terminated_at", service.GetTerminatedAt().UTC().String())
	d.Set("updated_at", service.GetUpdatedAt().UTC().String())
	d.Set("created_at", service.GetCreatedAt().UTC().String())

	return nil
}

func resourceKoyebServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	resolver := newIDResolver(client)
	var appId string

	if d.Get("app_name").(string) != "" {
		id, err := resolver.App(ctx, d.Get("app_name").(string))
		if err != nil {
			return diag.Errorf("Error creating service: %s", err)
		}

		appId = id
	}

	definition := expandDeploymentDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{}))

	res, resp, err := client.ServicesApi.CreateService(ctx).Service(koyeb.CreateService{
		AppId:      &appId,
		Definition: definition,
	}).Execute()
	if err != nil {
		return diag.Errorf("Error creating service: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.Service.Id)
	log.Printf("[INFO] Created service name: %s", *res.Service.Name)

	// Apply should not report success while the service is still starting;
	// HEALTHY/DEGRADED is the usable set shared with the Python SDK.
	if err := waitForServiceReady(ctx, client, d.Id(), serviceReadinessTimeout); err != nil {
		return diag.Errorf("Error waiting for service to be ready: %s", err)
	}

	return resourceKoyebServiceRead(ctx, d, meta)
}

func resourceKoyebServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	resolver := newIDResolver(client)
	var serviceId string

	if d.Id() != "" {
		id, err := resolver.Service(ctx, d.Id())
		if err != nil {
			return diag.Errorf("Error retrieving service: %s", err)
		}

		serviceId = id
	}

	serviceRes, resp, err := client.ServicesApi.GetService(ctx, serviceId).Execute()
	if err != nil {
		// If the service is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving service: %s (%v %v)", err, resp, serviceRes)
	}

	// deploymentRes, resp, err := client.DeploymentsApi.GetDeployment(ctx, *serviceRes.Service.LatestDeploymentId).Execute()
	// if err != nil {
	// 	return diag.Errorf("Error retrieving service latest deployment: %s (%v %v", err, resp, serviceRes)
	// }

	setServiceAttribute(d, serviceRes.Service)

	return nil
}

func resourceKoyebServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	definition := expandDeploymentDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{}))
	res, resp, err := client.ServicesApi.UpdateService(ctx, d.Id()).Service(koyeb.UpdateService{
		Definition: definition,
	}).Execute()
	if err != nil {
		return diag.Errorf("Error updating service: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Updated service name: %s", *res.Service.Name)

	if err := waitForServiceReady(ctx, client, d.Id(), serviceReadinessTimeout); err != nil {
		return diag.Errorf("Error waiting for service to be ready: %s", err)
	}

	return resourceKoyebServiceRead(ctx, d, meta)

}

func resourceKoyebServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.ServicesApi.DeleteService(ctx, d.Id()).Execute()
	if err != nil {
		return diag.Errorf("Error deleting service: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
