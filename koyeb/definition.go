package koyeb

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func deploymentDefinitionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The service name",
				ValidateFunc: validation.StringLenBetween(3, 64),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.EqualFold(old, new)
				},
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "WEB",
				// The server requires the definition name for SANDBOX pools;
				// `name` is Required above, which covers it.
				Description:  "The service type, either WEB, WORKER, DATABASE or SANDBOX (default WEB)",
				ValidateFunc: validation.StringInSlice([]string{"WEB", "WORKER", "DATABASE", "SANDBOX"}, false),
			},
			"docker": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     dockerSchema(),
				Set:      schema.HashResource(dockerSchema()),
				MaxItems: 1,
			},
			"archive": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "The archive to deploy, as uploaded by `koyeb deploy`",
				Elem:        archiveSourceSchema(),
				Set:         schema.HashResource(archiveSourceSchema()),
			},
			"git": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     gitSchema(),
				Set:      schema.HashResource(gitSchema()),
				MaxItems: 1,
			},
			"env": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     envSchema(),
				Set:      schema.HashResource(envSchema()),
			},
			"ports": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     portSchema(),
				Set:      schema.HashResource(portSchema()),
			},
			"proxy_ports": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The proxy ports to expose on the service (available for services of type WEB and SANDBOX)",
				Elem:        proxyPortSchema(),
				Set:         schema.HashResource(proxyPortSchema()),
			},
			"skip_cache": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If set to true, the service will be deployed without using the cache",
			},
			"strategy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "DEPLOYMENT_STRATEGY_TYPE_ROLLING",
				Description: "The deployment strategy used when updating the service: DEPLOYMENT_STRATEGY_TYPE_ROLLING, DEPLOYMENT_STRATEGY_TYPE_BLUE_GREEN or DEPLOYMENT_STRATEGY_TYPE_IMMEDIATE",
				ValidateFunc: validation.StringInSlice([]string{
					"DEPLOYMENT_STRATEGY_TYPE_ROLLING",
					"DEPLOYMENT_STRATEGY_TYPE_BLUE_GREEN",
					"DEPLOYMENT_STRATEGY_TYPE_IMMEDIATE",
				}, false),
			},
			"mesh": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "DEPLOYMENT_MESH_AUTO",
				Description: "Whether the service joins the service mesh: DEPLOYMENT_MESH_AUTO, DEPLOYMENT_MESH_ENABLED or DEPLOYMENT_MESH_DISABLED",
				ValidateFunc: validation.StringInSlice([]string{
					"DEPLOYMENT_MESH_AUTO",
					"DEPLOYMENT_MESH_ENABLED",
					"DEPLOYMENT_MESH_DISABLED",
				}, false),
			},
			"network_policy": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "The network policy applied to the service",
				Elem:        networkPolicySchema(),
				Set:         schema.HashResource(networkPolicySchema()),
			},
			"health_checks": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     healthCheckSchema(),
				Set:      schema.HashResource(healthCheckSchema()),
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
				Elem:     instanceTypeSchema(),
				Set:      schema.HashResource(instanceTypeSchema()),
			},
			"scalings": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     scalingSchema(),
				Set:      schema.HashResource(scalingSchema()),
			},
			"regions": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The service deployment regions to deploy to",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"volumes": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The volumes to attach and mount to the service",
				Elem:        serviceVolumeSchema(),
				Set:         schema.HashResource(serviceVolumeSchema()),
			},
			"config_files": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The config files to mount in the service",
				Elem:        configFileSchema(),
				Set:         schema.HashResource(configFileSchema()),
			},
			"database": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "The database to provision for the service (only for services of type DATABASE)",
				Elem:        databaseSourceSchema(),
				Set:         schema.HashResource(databaseSourceSchema()),
			},
		},
	}
}

func expandRegions(regions []interface{}) []string {
	expandedRegions := make([]string, len(regions))
	for i, v := range regions {
		expandedRegions[i] = v.(string)
	}

	return expandedRegions
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
		Name:          toOpt(rawDeploymentDefinition["name"].(string)),
		Type:          toOpt(koyeb.DeploymentDefinitionType(rawDeploymentDefinition["type"].(string))),
		Env:           expandEnvs(rawDeploymentDefinition["env"].(*schema.Set).List()),
		Ports:         expandPorts(rawDeploymentDefinition["ports"].(*schema.Set).List()),
		Routes:        expandRoutes(rawDeploymentDefinition["routes"].(*schema.Set).List()),
		Scalings:      expandScalings(rawDeploymentDefinition["scalings"].(*schema.Set).List()),
		InstanceTypes: expandInstanceTypes(rawDeploymentDefinition["instance_types"].(*schema.Set).List()),
		Regions:       expandRegions(rawDeploymentDefinition["regions"].(*schema.Set).List()),
		HealthChecks:  expandHealthChecks(rawDeploymentDefinition["health_checks"].(*schema.Set).List()),
		Volumes:       expandVolumes(rawDeploymentDefinition["volumes"].(*schema.Set).List()),
		ProxyPorts:    expandProxyPorts(rawDeploymentDefinition["proxy_ports"].(*schema.Set).List()),
		ConfigFiles:   expandConfigFiles(rawDeploymentDefinition["config_files"].(*schema.Set).List()),
	}

	database := rawDeploymentDefinition["database"].(*schema.Set).List()
	if len(database) > 0 {
		deploymentDefinition.Database = expandDatabaseSource(database)
	}

	networkPolicy := rawDeploymentDefinition["network_policy"].(*schema.Set).List()
	if len(networkPolicy) > 0 {
		deploymentDefinition.NetworkPolicy = expandNetworkPolicy(networkPolicy)
	}

	archive := rawDeploymentDefinition["archive"].(*schema.Set).List()
	if len(archive) > 0 {
		deploymentDefinition.Archive = expandArchiveSource(archive)
	}

	if strategy, ok := rawDeploymentDefinition["strategy"].(string); ok && strategy != "" {
		deploymentDefinition.Strategy = &koyeb.DeploymentStrategy{
			Type: toOpt(koyeb.DeploymentStrategyType(strategy)),
		}
	}
	if mesh, ok := rawDeploymentDefinition["mesh"].(string); ok && mesh != "" {
		deploymentDefinition.Mesh = toOpt(koyeb.DeploymentMesh(mesh))
	}

	git := rawDeploymentDefinition["git"].(*schema.Set).List()
	if len(git) > 0 {
		deploymentDefinition.Git = expandGitSource(git)
	}

	docker := rawDeploymentDefinition["docker"].(*schema.Set).List()
	if len(docker) > 0 {
		deploymentDefinition.Docker = expandDockerSource(docker)
	}

	return deploymentDefinition
}

func flattenDeploymentDefinition(deployment *koyeb.DeploymentDefinition) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["name"] = deployment.GetName()
	r["type"] = deployment.GetType()
	if docker, ok := deployment.GetDockerOk(); ok && docker != nil {
		r["docker"] = flattenDocker(docker)
	}
	if git, ok := deployment.GetGitOk(); ok && git != nil {
		r["git"] = flattenGit(git)
	}
	r["env"] = flattenEnvs(toOpt(deployment.GetEnv()))
	r["ports"] = flattenPorts(toOpt(deployment.GetPorts()))
	r["skip_cache"] = deployment.GetSkipCache()
	if strategy, ok := deployment.GetStrategyOk(); ok {
		if strategyType, ok := strategy.GetTypeOk(); ok {
			r["strategy"] = string(*strategyType)
		}
	}
	if mesh, ok := deployment.GetMeshOk(); ok {
		r["mesh"] = string(*mesh)
	}
	if check, ok := deployment.GetHealthChecksOk(); ok {
		r["health_checks"] = flattenHealthChecks(toOpt(check))
	}
	r["routes"] = flattenRoutes(toOpt(deployment.GetRoutes()))
	r["instance_types"] = flattenInstanceTypes(toOpt(deployment.GetInstanceTypes()))
	r["scalings"] = flattenScalings(toOpt(deployment.GetScalings()))
	r["regions"] = flattenRegions(&deployment.Regions)
	r["volumes"] = flattenVolumes(&deployment.Volumes)
	r["proxy_ports"] = flattenProxyPorts(&deployment.ProxyPorts)
	r["config_files"] = flattenConfigFiles(&deployment.ConfigFiles)
	if database, ok := deployment.GetDatabaseOk(); ok {
		r["database"] = flattenDatabase(database)
	}
	if networkPolicy, ok := deployment.GetNetworkPolicyOk(); ok {
		r["network_policy"] = flattenNetworkPolicy(networkPolicy)
	}
	if archive, ok := deployment.GetArchiveOk(); ok {
		r["archive"] = flattenArchive(archive)
	}

	result = append(result, r)

	return result
}

// Readiness budget for created and updated services; a variable so tests
// can shorten it.
