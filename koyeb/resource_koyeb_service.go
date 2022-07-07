package koyeb

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func envSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the environment variable",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The value of the environment variable",
				Sensitive:   true,
			},
		},
	}
}

func expandEnvs(config []interface{}) *[]koyeb.DeploymentEnv {
	envs := make([]koyeb.DeploymentEnv, 0, len(config))

	for _, rawEnv := range config {
		env := rawEnv.(map[string]interface{})

		e := koyeb.DeploymentEnv{
			Key: Ptr(env["key"].(string)),
		}

		if env["value"] != nil {
			e.Value = Ptr(env["value"].(string))
		}
		if env["secret"] != nil {
			e.Value = Ptr(env["secret"].(string))
		}

		envs = append(envs, e)
	}

	return &envs
}

func flattenEnvs(envs *[]koyeb.DeploymentEnv) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*envs))

	for i, env := range *envs {
		r := make(map[string]interface{})

		r["key"] = *env.Key

		if value, ok := env.GetValueOk(); ok {
			r["value"] = value
		}
		if secret, ok := env.GetSecretOk(); ok {
			r["secret"] = secret
		}

		result[i] = r
	}

	return result
}

func portSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The internal port on which this service's run command will listen",
			},
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The protocol used by your service",
				ValidateFunc: validation.StringInSlice([]string{
					"http",
					"tcp",
				}, false),
			},
		},
	}
}

func expandPorts(config []interface{}) *[]koyeb.DeploymentPort {
	ports := make([]koyeb.DeploymentPort, 0, len(config))

	for _, rawPort := range config {
		port := rawPort.(map[string]interface{})

		p := koyeb.DeploymentPort{
			Port:     Ptr(int64(port["port"].(int))),
			Protocol: Ptr(port["protocol"].(string)),
		}

		ports = append(ports, p)
	}

	return &ports
}

func flattenPorts(ports *[]koyeb.DeploymentPort) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*ports))

	for i, port := range *ports {
		r := make(map[string]interface{})

		r["port"] = *port.Port
		r["protocol"] = *port.Protocol

		result[i] = r
	}

	return result
}

func routeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The internal port on which this service's run command will listen",
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path specifies an route by HTTP path prefix. Paths must start with / and must be unique within the app",
				Sensitive:   true,
			},
		},
	}
}

func expandRoutes(config []interface{}) *[]koyeb.DeploymentRoute {
	routes := make([]koyeb.DeploymentRoute, 0, len(config))

	for _, rawRoute := range config {
		route := rawRoute.(map[string]interface{})

		r := koyeb.DeploymentRoute{
			Port: Ptr(int64(route["port"].(int))),
			Path: Ptr(route["path"].(string)),
		}

		routes = append(routes, r)
	}

	return &routes
}

func flattenRoutes(routes *[]koyeb.DeploymentRoute) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*routes))

	for i, route := range *routes {
		r := make(map[string]interface{})

		r["port"] = route.Port
		r["path"] = route.Path

		result[i] = r
	}

	return result
}

func instanceTypeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The instance type to use to support your service",
			},
		},
	}
}

func expandInstanceTypes(config []interface{}) *[]koyeb.DeploymentInstanceType {
	instanceTypes := make([]koyeb.DeploymentInstanceType, 0, len(config))

	for _, rawInstanceType := range config {
		instanceType := rawInstanceType.(map[string]interface{})

		r := koyeb.DeploymentInstanceType{
			Type: Ptr(instanceType["type"].(string)),
		}

		instanceTypes = append(instanceTypes, r)
	}

	return &instanceTypes
}

func flattenInstanceTypes(instanceTypes *[]koyeb.DeploymentInstanceType) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*instanceTypes))

	for i, instanceType := range *instanceTypes {
		r := make(map[string]interface{})

		r["type"] = instanceType.Type

		result[i] = r
	}

	return result
}

func scalingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"min": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The instance type to use to support your service",
			},
			"max": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The instance type to use to support your service",
			},
		},
	}
}

func expandScalings(config []interface{}) *[]koyeb.DeploymentScaling {
	scalings := make([]koyeb.DeploymentScaling, 0, len(config))
	diag.Errorf("Error updating secret: %v", config)
	for _, rawScaling := range config {
		scaling := rawScaling.(map[string]interface{})

		r := koyeb.DeploymentScaling{
			Max: Ptr(int64(scaling["max"].(int))),
			Min: Ptr(int64(scaling["min"].(int))),
		}

		scalings = append(scalings, r)
	}

	return &scalings
}

func flattenScalings(scalings *[]koyeb.DeploymentScaling) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*scalings))

	for i, scaling := range *scalings {
		r := make(map[string]interface{})

		r["max"] = *scaling.Max
		r["min"] = *scaling.Min

		result[i] = r
	}

	return result
}

func dockerSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Docker image to use to support your service",
			},
		},
	}
}

func expandDockerSource(config []interface{}) *koyeb.DockerSource {
	rawDockerSource := config[0].(map[string]interface{})

	dockerSource := &koyeb.DockerSource{
		Image: Ptr(rawDockerSource["image"].(string)),
	}

	return dockerSource
}

func flattenDocker(dockerSource *koyeb.DockerSource) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["image"] = dockerSource.Image

	result = append(result, r)

	return result
}

func deploymentDefinitionSchena() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The service name",
				ValidateFunc: validation.StringLenBetween(3, 64),
			},
			"docker": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     dockerSchema(),
				Set:      schema.HashResource(dockerSchema()),
			},
			"env": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     envSchema(),
				Set:      schema.HashResource(envSchema()),
			},
			"ports": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     portSchema(),
				Set:      schema.HashResource(routeSchema()),
			},
			"routes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     routeSchema(),
				Set:      schema.HashResource(routeSchema()),
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     instanceTypeSchema(),
				Set:      schema.HashResource(instanceTypeSchema()),
			},
			"scalings": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     scalingSchema(),
				Set:      schema.HashResource(scalingSchema()),
			},
			"regions": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The service deployment regions to deploy to",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func expandRegions(regions []interface{}) *[]string {
	expandedRegions := make([]string, len(regions))
	for i, v := range regions {
		expandedRegions[i] = v.(string)
	}

	return &expandedRegions
}

func flattenRegions(regions *[]string) *schema.Set {
	flattenedRegions := schema.NewSet(schema.HashString, []interface{}{})
	for _, r := range *regions {
		flattenedRegions.Add(r)
	}

	return flattenedRegions
}

func expandDeploymentDefinition(configmap map[string]interface{}) *koyeb.DeploymentDefinition {
	rawDeploymentDefinition := configmap

	deploymentDefinition := &koyeb.DeploymentDefinition{
		Name:          Ptr(rawDeploymentDefinition["name"].(string)),
		Docker:        expandDockerSource(rawDeploymentDefinition["docker"].(*schema.Set).List()),
		Env:           expandEnvs(rawDeploymentDefinition["env"].(*schema.Set).List()),
		Ports:         expandPorts(rawDeploymentDefinition["ports"].(*schema.Set).List()),
		Routes:        expandRoutes(rawDeploymentDefinition["routes"].(*schema.Set).List()),
		Scalings:      expandScalings(rawDeploymentDefinition["scalings"].(*schema.Set).List()),
		InstanceTypes: expandInstanceTypes(rawDeploymentDefinition["instance_types"].(*schema.Set).List()),
		Regions:       expandRegions(rawDeploymentDefinition["regions"].(*schema.Set).List()),
	}

	return deploymentDefinition
}

func flattenDeploymentDefinition(deployment *koyeb.DeploymentDefinition) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["name"] = deployment.Name
	r["docker"] = flattenDocker(deployment.Docker)
	r["env"] = flattenEnvs(deployment.Env)
	r["ports"] = flattenPorts(deployment.Ports)
	r["routes"] = flattenRoutes(deployment.Routes)
	r["instance_types"] = flattenInstanceTypes(deployment.InstanceTypes)
	r["scalings"] = flattenScalings(deployment.Scalings)
	r["regions"] = flattenRegions(deployment.Regions)

	result = append(result, r)

	return result
}

func deploymentSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"definition": {
				Type:        schema.TypeSet,
				MinItems:    1,
				MaxItems:    1,
				Required:    true,
				Description: "The service deployment definition",
				Elem:        deploymentDefinitionSchena(),
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
			"child_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"parent_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"terminated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"succeeded_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"started_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was last updated",
			},
			"allocated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of when the service was created",
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

func flattenDeployment(deployment *koyeb.Deployment) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["id"] = deployment.GetId()
	r["definition"] = flattenDeploymentDefinition(deployment.Definition)
	r["version"] = deployment.GetVersion()
	r["status"] = deployment.GetStatus()
	r["messages"] = strings.Join(deployment.GetMessages(), " ")
	r["child_id"] = deployment.GetChildId()
	r["parent_id"] = deployment.GetParentId()
	r["terminated_at"] = deployment.GetTerminatedAt().UTC().String()
	r["succeeded_at"] = deployment.GetSucceededAt().UTC().String()
	r["started_at"] = deployment.GetStartedAt().UTC().String()
	r["allocated_at"] = deployment.GetAllocatedAt().UTC().String()
	r["updated_at"] = deployment.GetUpdatedAt().UTC().String()
	r["created_at"] = deployment.GetCreatedAt().UTC().String()

	result = append(result, r)

	return result
}

func serviceSchema() map[string]*schema.Schema {
	service := map[string]*schema.Schema{
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
		"app_name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "The app name the service is assigned",
			ValidateFunc: validation.StringLenBetween(3, 23),
		},
		"app_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The app id the service is assigned",
		},
		"definition": {
			Type:        schema.TypeSet,
			MinItems:    1,
			MaxItems:    1,
			Required:    true,
			Description: "The service deployment definition",
			Elem:        deploymentDefinitionSchena(),
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
	}

	return service
}

func flattenService(service *koyeb.Service) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["id"] = service.GetId()
	r["name"] = service.GetName()
	r["paused_at"] = service.GetCreatedAt().UTC().String()
	r["resumed_at"] = service.GetCreatedAt().UTC().String()
	r["terminated_at"] = service.GetCreatedAt().UTC().String()
	r["created_at"] = service.GetCreatedAt().UTC().String()
	r["updated_at"] = service.GetUpdatedAt().UTC().String()

	return result
}

func resourceKoyebService() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Service resource in the Koyeb Terraform provider.",

		CreateContext: resourceKoyebServiceCreate,
		ReadContext:   resourceKoyebServiceRead,
		UpdateContext: resourceKoyebServiceUpdate,
		DeleteContext: resourceKoyebServiceDelete,

		Schema: serviceSchema(),
	}
}

func setServiceAttribute(
	d *schema.ResourceData,
	service *koyeb.Service,
	// activeDeployment *koyeb.Deployment,
	// latestDeployment *koyeb.Deployment,
) error {
	d.SetId(service.GetId())
	d.Set("id", service.GetId())
	d.Set("name", service.GetName())
	d.Set("app_id", service.GetAppId())
	d.Set("version", service.GetVersion())
	d.Set("status", service.GetStatus())
	d.Set("messages", strings.Join(service.GetMessages(), " "))
	d.Set("paused_at", service.GetPausedAt().UTC().String())
	d.Set("resumed_at", service.GetResumedAt().UTC().String())
	d.Set("terminated_at", service.GetTerminatedAt().UTC().String())
	d.Set("created_at", service.GetCreatedAt().UTC().String())
	d.Set("updated_at", service.GetUpdatedAt().UTC().String())
	d.Set("latest_deployment", service.GetLatestDeploymentId())
	d.Set("active_deployment", service.GetActiveDeploymentId())
	d.Set("organization_id", service.GetOrganizationId())

	// if _, ok := activeDeployment.GetIdOk(); ok {
	// 	d.Set("active_deployment", flattenDeployment(activeDeployment))
	// }

	// if _, ok := latestDeployment.GetIdOk(); ok {
	// 	d.Set("latest_deployment", flattenDeployment(latestDeployment))
	// }

	return nil
}

func resourceKoyebServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	mapper := idmapper.NewMapper(ctx, client)
	appMapper := mapper.App()
	var appId string

	if d.Get("app_name").(string) != "" {
		id, err := appMapper.ResolveID(d.Get("app_name").(string))

		if err != nil {
			return diag.Errorf("Error creating domain: %s", err)
		}

		appId = id
	}

	definition := expandDeploymentDefinition(d.Get("definition").(*schema.Set).List()[0].(map[string]interface{}))

	res, resp, err := client.ServicesApi.CreateService(ctx).Body(koyeb.CreateService{
		AppId:      &appId,
		Definition: definition,
	}).Execute()
	if err != nil {
		return diag.Errorf("Error creating service: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.Service.Id)
	log.Printf("[INFO] Created service name: %s", *res.Service.Name)

	return resourceKoyebServiceRead(ctx, d, meta)
}

func resourceKoyebServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	// var activeDeployment *koyeb.Deployment
	// var latestDeployment *koyeb.Deployment

	res, resp, err := client.ServicesApi.GetService(ctx, d.Id()).Execute()
	if err != nil {
		// If the service is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving service: %s (%v %v)", err, resp, res)
	}

	// if activeDeploymentId, ok := res.Service.GetActiveDeploymentIdOk(); ok {
	// 	res, resp, err := client.DeploymentsApi.GetDeployment(ctx, *activeDeploymentId).Execute()
	// 	if err != nil {
	// 		return diag.Errorf("Error retrieving service active deploymen (%s)t:  (%v %v)", *activeDeploymentId, err, resp, res)
	// 	}

	// 	activeDeployment = res.Deployment
	// }

	// if latestDeploymentId, ok := res.Service.GetLatestDeploymentIdOk(); ok {
	// 	res, resp, err := client.DeploymentsApi.GetDeployment(ctx, *latestDeploymentId).Execute()
	// 	if err != nil {
	// 		return diag.Errorf("Error retrieving service active deployment: %s (%v %v", err, resp, res)
	// 	}

	// 	latestDeployment = res.Deployment
	// }

	// err = setServiceAttribute(d, res.Service, activeDeployment, latestDeployment)
	err = setServiceAttribute(d, res.Service)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKoyebServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	definition := expandDeploymentDefinition(d.Get("definition").(*schema.Set).List()[0].(map[string]interface{}))

	res, resp, err := client.ServicesApi.UpdateService(ctx, d.Id()).Body(koyeb.UpdateService{
		Definition: definition,
	}).Execute()

	if err != nil {
		return diag.Errorf("Error updating service: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Updated service name: %s", *res.Service.Name)
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
