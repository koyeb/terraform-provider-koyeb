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
			Elem:        deploymentDefinitionSchena(),
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

func deploymentDefinitionSchena() *schema.Resource {
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "WEB",
				Description:  "The service type, either WEB or WORKER (default WEB)",
				ValidateFunc: validation.StringInSlice([]string{"WEB", "WORKER"}, false),
			},
			"docker": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     dockerSchema(),
				Set:      schema.HashResource(dockerSchema()),
				MaxItems: 1,
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
			"skip_cache": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If set to true, the service will be deployed without using the cache",
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
		},
	}
}

func dockerSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Docker image to use to support your service",
			},
			"command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Docker command to use",
			},
			"args": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Docker args to use",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"entrypoint": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Docker entrypoint to use",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"privileged": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When enabled, the service container will run in privileged mode. This advanced feature is useful to get advanced system privileges.",
			},
			"image_registry_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Koyeb secret containing the container registry credentials",
			},
		},
	}
}

func gitSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The GitHub repository to deploy",
			},
			"branch": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The GitHub branch to deploy",
			},
			"workdir": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The directory where your source code is located. If not set, the work directory defaults to the root of the repository.",
			},
			"buildpack": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     buildpackBuilderSchema(),
				Set:      schema.HashResource(buildpackBuilderSchema()),
				MaxItems: 1,
			},
			"dockerfile": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     dockerBuilderSchema(),
				Set:      schema.HashResource(dockerBuilderSchema()),
				MaxItems: 1,
			},
			"no_deploy_on_push": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to true, no Koyeb deployments will be triggered when changes are pushed to the GitHub repository branch",
			},
		},
	}
}

func buildpackBuilderSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"build_command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The command to build your application during the build phase. If your application does not require a build command, leave this field empty",
			},
			"run_command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The command to run your application once the built is completed",
			},
			"privileged": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When enabled, the service container will run in privileged mode. This advanced feature is useful to get advanced system privileges.",
			},
		},
	}
}

func dockerBuilderSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dockerfile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The location of your Dockerfile relative to the work directory. If not set, the work directory defaults to the root of the repository.",
			},
			"entrypoint": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Override the default entrypoint to execute on the container",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Override the command to execute on the container",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"args": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The arguments to pass to the Docker command",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"target": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Target build stage: If your Dockerfile contains multi-stage builds, you can choose the target stage to build and deploy by entering its name",
			},
			"privileged": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "When enabled, the service container will run in privileged mode. This advanced feature is useful to get advanced system privileges.",
			},
		},
	}
}

func envSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scopes": {
				Type:     schema.TypeList,
				Optional: true,
				// Computed:    true,
				Description: "The regions the environment variable needs to be exposed",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the environment variable",
			},
			"value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The value of the environment variable",
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The secret name to use as the value of the environment variable",
			},
		},
	}
}

func portSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The internal port on which this service's run command will listen",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"protocol": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The protocol used by your service",
				ValidateFunc: validation.StringInSlice([]string{
					"http",
					"http2",
					"tcp",
				}, false),
			},
		},
	}
}

func healthCheckSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"grace_period": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The period in seconds to wait for the instance to become healthy, default is 5s",
			},
			"interval": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The period in seconds between two health checks, default is 60s",
			},
			"restart_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The number of consecutive failures before attempting to restart the service, default is 3",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum time to wait in seconds before considering the check as a failure, default is 5s",
			},
			"tcp": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     TCPHealthCheckSchema(),
				Set:      schema.HashResource(TCPHealthCheckSchema()),
				MaxItems: 1,
			},
			"http": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     HTTPHealthCheckSchema(),
				Set:      schema.HashResource(HTTPHealthCheckSchema()),
				MaxItems: 1,
			},
		},
	}
}

func TCPHealthCheckSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The port to use to perform the health check",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
		},
	}
}

func HTTPHealthCheckSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The port to use to perform the health check",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path to use to perform the HTTP health check",
			},
			"method": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional HTTP method to use to perform the health check, default is GET",
			},
			"headers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     HTTPHealthCheckHeaderSchema(),
				Set:      schema.HashResource(HTTPHealthCheckHeaderSchema()),
			},
		},
	}
}

func HTTPHealthCheckHeaderSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the header",
			},
			"value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The value of the header",
			},
		},
	}
}

func routeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The internal port on which this service's run command will listen",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Path specifies a route by HTTP path prefix. Paths must start with / and must be unique within the app",
			},
		},
	}
}

func instanceTypeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scopes": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The regions to use the instance type",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The instance type to use to support your service",
			},
		},
	}
}

func scalingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scopes": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The regions to apply the scaling configuration",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"min": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The minimal number of instances to use to support your service",
			},
			"max": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "The maximum number of instance to use to support your service",
			},
			"targets": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     autoScalingTargetSchema(),
				Set:      schema.HashResource(autoScalingTargetSchema()),
			},
		},
	}
}

func autoScalingTargetSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"average_cpu": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The CPU usage (expressed as a percentage) across all Instances of your Service within a region",
				Elem:        autoScalingTargetValueSchema(),
				Set:         schema.HashResource(autoScalingTargetValueSchema()),
			},
			"average_mem": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The memory usage (expressed as a percentage) across all Instances of your Service within a region",
				Elem:        autoScalingTargetValueSchema(),
				Set:         schema.HashResource(autoScalingTargetValueSchema()),
			},
			"requests_per_second": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The number of concurrent requests per second across all Instances of your Service within a region",
				Elem:        autoScalingTargetValueSchema(),
				Set:         schema.HashResource(autoScalingTargetValueSchema()),
			},
			"concurrent_requests": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The number of concurrent requests across all Instances of your Service within a region",
				Elem:        autoScalingTargetValueSchema(),
				Set:         schema.HashResource(autoScalingTargetValueSchema()),
			},
			"request_response_time": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The average response time of requests across all Instances of your Service within a region",
				Elem:        autoScalingTargetValueSchema(),
				Set:         schema.HashResource(autoScalingTargetValueSchema()),
			},
		},
	}
}

func autoScalingTargetValueSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The target value of the autoscaling target",
			},
		},
	}
}

func serviceVolumeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scope": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The regions to apply the scaling configuration",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The volume ID to mount to the service",
			},
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The path where to mount the volume",
			},
			"replica_index": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Explicitly specify the replica index to mount the volume to",
			},
		},
	}
}

func expandEnvs(config []interface{}) []koyeb.DeploymentEnv {
	envs := make([]koyeb.DeploymentEnv, 0, len(config))

	for _, rawEnv := range config {
		env := rawEnv.(map[string]interface{})

		e := koyeb.DeploymentEnv{
			Key: toOpt(env["key"].(string)),
		}

		rawScopes := env["scopes"].([]interface{})
		scopes := make([]string, len(rawScopes))
		for i, v := range rawScopes {
			scopes[i] = v.(string)
		}
		e.Scopes = scopes

		if env["value"] != nil && env["value"].(string) != "" {
			e.Value = toOpt(env["value"].(string))
		}
		if env["secret"] != nil && env["secret"].(string) != "" {
			e.Secret = toOpt(env["secret"].(string))
		}

		envs = append(envs, e)
	}

	return envs
}

func flattenEnvs(envs *[]koyeb.DeploymentEnv) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*envs))

	for i, env := range *envs {
		r := make(map[string]interface{})

		r["key"] = env.GetKey()
		// r["scopes"] = env.GetScopes()

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

func expandPorts(config []interface{}) []koyeb.DeploymentPort {
	ports := make([]koyeb.DeploymentPort, 0, len(config))

	for _, rawPort := range config {
		port := rawPort.(map[string]interface{})

		p := koyeb.DeploymentPort{
			Port:     toOpt(int64(port["port"].(int))),
			Protocol: toOpt(port["protocol"].(string)),
		}

		ports = append(ports, p)
	}

	return ports
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

func expandRoutes(config []interface{}) []koyeb.DeploymentRoute {
	routes := make([]koyeb.DeploymentRoute, 0, len(config))

	for _, rawRoute := range config {
		route := rawRoute.(map[string]interface{})

		r := koyeb.DeploymentRoute{
			Port: toOpt(int64(route["port"].(int))),
			Path: toOpt(route["path"].(string)),
		}

		routes = append(routes, r)
	}

	return routes
}

func flattenRoutes(routes *[]koyeb.DeploymentRoute) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*routes))

	for i, route := range *routes {
		r := make(map[string]interface{})

		r["port"] = route.GetPort()
		r["path"] = route.GetPath()

		result[i] = r
	}

	return result
}

func expandInstanceTypes(config []interface{}) []koyeb.DeploymentInstanceType {
	instanceTypes := make([]koyeb.DeploymentInstanceType, 0, len(config))

	for _, rawInstanceType := range config {
		instanceType := rawInstanceType.(map[string]interface{})

		r := koyeb.DeploymentInstanceType{
			Type: toOpt(instanceType["type"].(string)),
		}

		rawScopes := instanceType["scopes"].([]interface{})
		scopes := make([]string, len(rawScopes))
		for i, v := range rawScopes {
			scopes[i] = v.(string)
		}
		r.Scopes = scopes

		instanceTypes = append(instanceTypes, r)
	}

	return instanceTypes
}

func flattenInstanceTypes(instanceTypes *[]koyeb.DeploymentInstanceType) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*instanceTypes))

	for i, instanceType := range *instanceTypes {
		r := make(map[string]interface{})

		r["type"] = instanceType.GetType()
		// r["scopes"] = instanceType.GetScopes()

		result[i] = r
	}

	return result
}

func expandScalings(config []interface{}) []koyeb.DeploymentScaling {
	scalings := make([]koyeb.DeploymentScaling, 0, len(config))

	for _, rawScalings := range config {
		scaling := rawScalings.(map[string]interface{})

		s := koyeb.DeploymentScaling{
			Max: toOpt(int64(scaling["max"].(int))),
			Min: toOpt(int64(scaling["min"].(int))),
		}

		rawScopes := scaling["scopes"].([]interface{})
		scopes := make([]string, len(rawScopes))
		for i, v := range rawScopes {
			scopes[i] = v.(string)
		}
		s.Scopes = scopes

		targets := scaling["targets"].(*schema.Set).List()
		for _, rawTarget := range targets {
			target := rawTarget.(map[string]interface{})

			if target["average_cpu"] != nil {
				cpu := target["average_cpu"].(*schema.Set).List()
				for _, rawCPU := range cpu {
					cpu := rawCPU.(map[string]interface{})
					s.Targets = append(s.Targets, koyeb.DeploymentScalingTarget{
						AverageCpu: &koyeb.DeploymentScalingTargetAverageCPU{
							Value: toOpt(int64(cpu["value"].(int))),
						},
					})
				}
			}
			if target["average_mem"] != nil {
				mem := target["average_mem"].(*schema.Set).List()
				for _, rawMem := range mem {
					mem := rawMem.(map[string]interface{})
					s.Targets = append(s.Targets, koyeb.DeploymentScalingTarget{
						AverageMem: &koyeb.DeploymentScalingTargetAverageMem{
							Value: toOpt(int64(mem["value"].(int))),
						},
					})
				}
			}

			if target["requests_per_second"] != nil {
				rps := target["requests_per_second"].(*schema.Set).List()
				for _, rawRPS := range rps {
					rps := rawRPS.(map[string]interface{})
					s.Targets = append(s.Targets, koyeb.DeploymentScalingTarget{
						RequestsPerSecond: &koyeb.DeploymentScalingTargetRequestsPerSecond{
							Value: toOpt(int64(rps["value"].(int))),
						},
					})
				}
			}

			if target["concurrent_requests"] != nil {
				concReq := target["concurrent_requests"].(*schema.Set).List()
				for _, rawConcReq := range concReq {
					concReq := rawConcReq.(map[string]interface{})
					s.Targets = append(s.Targets, koyeb.DeploymentScalingTarget{
						ConcurrentRequests: &koyeb.DeploymentScalingTargetConcurrentRequests{
							Value: toOpt(int64(concReq["value"].(int))),
						},
					})
				}
			}

			if target["request_response_time"] != nil {
				reqRespTime := target["request_response_time"].(*schema.Set).List()
				for _, rawReqRespTime := range reqRespTime {
					reqRespTime := rawReqRespTime.(map[string]interface{})
					s.Targets = append(s.Targets, koyeb.DeploymentScalingTarget{
						RequestsResponseTime: &koyeb.DeploymentScalingTargetRequestsResponseTime{
							Value: toOpt(int64(reqRespTime["value"].(int))),
						},
					})
				}
			}

		}

		scalings = append(scalings, s)
	}

	return scalings
}

func flattenScalings(scalings *[]koyeb.DeploymentScaling) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*scalings))

	for i, scaling := range *scalings {
		r := make(map[string]interface{})

		r["max"] = scaling.GetMax()
		r["min"] = scaling.GetMin()
		// r["scopes"] = scaling.GetScopes()

		targetMap := make(map[string]interface{})
		for _, target := range scaling.Targets {

			if cpu, ok := target.GetAverageCpuOk(); ok {
				targetMap["average_cpu"] = schema.NewSet(
					schema.HashResource(autoScalingTargetValueSchema()),
					[]interface{}{
						map[string]interface{}{
							"value": int(cpu.GetValue()),
						},
					},
				)
			}
			if mem, ok := target.GetAverageMemOk(); ok {
				targetMap["average_mem"] = schema.NewSet(
					schema.HashResource(autoScalingTargetValueSchema()),
					[]interface{}{
						map[string]interface{}{
							"value": int(mem.GetValue()),
						},
					},
				)
			}
			if rps, ok := target.GetRequestsPerSecondOk(); ok {
				targetMap["requests_per_second"] = schema.NewSet(
					schema.HashResource(autoScalingTargetValueSchema()),
					[]interface{}{
						map[string]interface{}{
							"value": int(rps.GetValue()),
						},
					},
				)
			}
			if concReq, ok := target.GetConcurrentRequestsOk(); ok {
				targetMap["concurrent_requests"] = schema.NewSet(
					schema.HashResource(autoScalingTargetValueSchema()),
					[]interface{}{
						map[string]interface{}{
							"value": int(concReq.GetValue()),
						},
					},
				)
			}
			if reqRespTime, ok := target.GetRequestsResponseTimeOk(); ok {
				targetMap["request_response_time"] = schema.NewSet(
					schema.HashResource(autoScalingTargetValueSchema()),
					[]interface{}{
						map[string]interface{}{
							"value": int(reqRespTime.GetValue()),
						},
					},
				)
			}

		}
		r["targets"] = schema.NewSet(
			schema.HashResource(autoScalingTargetSchema()),
			[]interface{}{targetMap},
		)
		result[i] = r
	}

	return result
}

func expandDockerSource(config []interface{}) *koyeb.DockerSource {
	rawDockerSource := config[0].(map[string]interface{})

	dockerSource := &koyeb.DockerSource{
		Image: toOpt(rawDockerSource["image"].(string)),
	}

	if rawDockerSource["command"] != nil {
		dockerSource.Command = toOpt(rawDockerSource["command"].(string))
	}

	rawArgs := rawDockerSource["args"].([]interface{})
	args := make([]string, len(rawArgs))
	for i, v := range rawArgs {
		args[i] = v.(string)
	}
	dockerSource.Args = args

	rawEntrypoint := rawDockerSource["entrypoint"].([]interface{})
	entrypoint := make([]string, len(rawEntrypoint))
	for i, v := range rawEntrypoint {
		entrypoint[i] = v.(string)
	}
	dockerSource.Entrypoint = entrypoint

	if rawDockerSource["privileged"] != nil {
		dockerSource.Privileged = toOpt(rawDockerSource["privileged"].(bool))
	}

	if rawDockerSource["image_registry_secret"] != nil {
		dockerSource.ImageRegistrySecret = toOpt(rawDockerSource["image_registry_secret"].(string))
	}

	return dockerSource
}

func flattenDocker(dockerSource *koyeb.DockerSource) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["image"] = dockerSource.Image
	r["command"] = dockerSource.Command
	r["args"] = dockerSource.Args
	r["entrypoint"] = dockerSource.Entrypoint
	r["privileged"] = dockerSource.Privileged
	r["image_registry_secret"] = dockerSource.ImageRegistrySecret

	result = append(result, r)

	return result
}

func expandDockerBuilder(config []interface{}) *koyeb.DockerBuilder {
	rawDockerBuilderSource := config[0].(map[string]interface{})

	dockerBuilderSource := &koyeb.DockerBuilder{}

	if rawDockerBuilderSource["dockerfile"] != nil {
		dockerBuilderSource.Dockerfile = toOpt(rawDockerBuilderSource["dockerfile"].(string))
	}

	rawEntrypoint := rawDockerBuilderSource["entrypoint"].([]interface{})
	entrypoint := make([]string, len(rawEntrypoint))
	for i, v := range rawEntrypoint {
		entrypoint[i] = v.(string)
	}
	dockerBuilderSource.Entrypoint = entrypoint

	if rawDockerBuilderSource["command"] != nil {
		dockerBuilderSource.Command = toOpt(rawDockerBuilderSource["command"].(string))
	}

	rawArgs := rawDockerBuilderSource["args"].([]interface{})
	args := make([]string, len(rawArgs))
	for i, v := range rawArgs {
		args[i] = v.(string)
	}
	dockerBuilderSource.Args = args

	if rawDockerBuilderSource["target"] != nil {
		dockerBuilderSource.Target = toOpt(rawDockerBuilderSource["target"].(string))
	}

	if rawDockerBuilderSource["privileged"] != nil {
		dockerBuilderSource.Privileged = toOpt(rawDockerBuilderSource["privileged"].(bool))
	}

	return dockerBuilderSource
}

func flattenDockerBuilder(dockerBuilderSource *koyeb.DockerBuilder) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["entrypoint"] = dockerBuilderSource.Entrypoint
	r["command"] = dockerBuilderSource.Command
	r["args"] = dockerBuilderSource.Args
	r["target"] = dockerBuilderSource.Target
	r["privileged"] = dockerBuilderSource.Privileged

	result = append(result, r)

	return result
}

func expandBuildpackBuilder(config []interface{}) *koyeb.BuildpackBuilder {
	rawBuildpackBuilderSource := config[0].(map[string]interface{})

	buildpackBuilderSource := &koyeb.BuildpackBuilder{}

	if rawBuildpackBuilderSource["build_command"] != nil {
		buildpackBuilderSource.BuildCommand = toOpt(rawBuildpackBuilderSource["build_command"].(string))
	}

	if rawBuildpackBuilderSource["run_command"] != nil {
		buildpackBuilderSource.RunCommand = toOpt(rawBuildpackBuilderSource["run_command"].(string))
	}

	if rawBuildpackBuilderSource["privileged"] != nil {
		buildpackBuilderSource.Privileged = toOpt(rawBuildpackBuilderSource["privileged"].(bool))
	}

	return buildpackBuilderSource
}

func flattenBuildpackBuilder(buildpackBuilderSource *koyeb.BuildpackBuilder) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["build_command"] = buildpackBuilderSource.GetBuildCommand()
	r["run_command"] = buildpackBuilderSource.GetRunCommand()
	r["privileged"] = buildpackBuilderSource.GetPrivileged()

	result = append(result, r)

	return result
}

func expandGitSource(config []interface{}) *koyeb.GitSource {
	rawGitSource := config[0].(map[string]interface{})

	gitSource := &koyeb.GitSource{
		Repository:     toOpt(rawGitSource["repository"].(string)),
		Branch:         toOpt(rawGitSource["branch"].(string)),
		Workdir:        toOpt(rawGitSource["workdir"].(string)),
		NoDeployOnPush: toOpt(rawGitSource["no_deploy_on_push"].(bool)),
	}

	if rawGitSource["dockerfile"] != nil && rawGitSource["dockerfile"].(*schema.Set).Len() > 0 {
		gitSource.Docker = expandDockerBuilder(rawGitSource["dockerfile"].(*schema.Set).List())
	} else if rawGitSource["buildpack"] != nil && rawGitSource["buildpack"].(*schema.Set).Len() > 0 {
		gitSource.Buildpack = expandBuildpackBuilder(rawGitSource["buildpack"].(*schema.Set).List())
	}

	return gitSource
}

func flattenGit(gitSource *koyeb.GitSource) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["repository"] = gitSource.GetRepository()
	r["branch"] = gitSource.GetBranch()
	r["workdir"] = gitSource.GetWorkdir()
	r["no_deploy_on_push"] = gitSource.GetNoDeployOnPush()
	if buildpack, ok := gitSource.GetBuildpackOk(); ok {
		r["buildpack"] = flattenBuildpackBuilder(buildpack)
	}
	if docker, ok := gitSource.GetDockerOk(); ok {
		r["dockerfile"] = flattenDockerBuilder(docker)
	}

	result = append(result, r)

	return result
}

func expandHealthChecks(config []interface{}) []koyeb.DeploymentHealthCheck {
	healthChecks := make([]koyeb.DeploymentHealthCheck, 0, len(config))

	for _, rawHealthCheck := range config {
		healthCheck := rawHealthCheck.(map[string]interface{})

		c := koyeb.DeploymentHealthCheck{
			GracePeriod:  toOpt(int64(healthCheck["grace_period"].(int))),
			Interval:     toOpt(int64(healthCheck["interval"].(int))),
			RestartLimit: toOpt(int64(healthCheck["restart_limit"].(int))),
			Timeout:      toOpt(int64(healthCheck["timeout"].(int))),
		}

		tcp := healthCheck["tcp"].(*schema.Set).List()
		if len(tcp) > 0 {
			tcphealthCheck := tcp[0].(map[string]interface{})

			c.Tcp = &koyeb.TCPHealthCheck{
				Port: toOpt(int64(tcphealthCheck["port"].(int))),
			}
		}

		http := healthCheck["http"].(*schema.Set).List()
		if len(http) > 0 {
			httpHealthCheck := http[0].(map[string]interface{})

			headers := make([]koyeb.HTTPHeader, 0, len(config))

			for _, rawHTTPHeader := range httpHealthCheck["headers"].(*schema.Set).List() {

				header := rawHTTPHeader.(map[string]interface{})

				h := koyeb.HTTPHeader{
					Key:   toOpt(header["key"].(string)),
					Value: toOpt(header["value"].(string)),
				}

				headers = append(headers, h)
			}

			c.Http = &koyeb.HTTPHealthCheck{
				Port:    toOpt(int64(httpHealthCheck["port"].(int))),
				Path:    toOpt(httpHealthCheck["path"].(string)),
				Headers: headers,
			}

			if httpHealthCheck["method"] != nil {
				c.Http.Method = toOpt(httpHealthCheck["method"].(string))
			}

		}

		healthChecks = append(healthChecks, c)
	}

	return healthChecks
}

func flattenHTTPHealthCheckHeaders(headers []koyeb.HTTPHeader) []map[string]interface{} {
	result := make([]map[string]interface{}, len(headers))

	for i, header := range headers {
		r := make(map[string]interface{})

		r["key"] = header.GetKey()
		r["value"] = header.GetValue()

		result[i] = r
	}

	return result
}

func flattenHealthChecks(healthChecks *[]koyeb.DeploymentHealthCheck) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*healthChecks))

	for i, check := range *healthChecks {
		r := make(map[string]interface{})

		r["grace_period"] = check.GetGracePeriod()
		r["interval"] = check.GetInterval()
		r["restart_limit"] = check.GetRestartLimit()
		r["timeout"] = check.GetTimeout()

		if tcp, ok := check.GetTcpOk(); ok {
			tcpEntry := map[string]interface{}{
				"port": int(tcp.GetPort()),
			}

			r["tcp"] = schema.NewSet(
				schema.HashResource(TCPHealthCheckSchema()),
				[]interface{}{tcpEntry},
			)
		}

		if http, ok := check.GetHttpOk(); ok {
			httpEntry := map[string]interface{}{
				"port":   int(http.GetPort()),
				"path":   http.GetPath(),
				"method": http.GetMethod(),
			}

			headers := flattenHTTPHealthCheckHeaders(http.GetHeaders())
			var headerInterfaces []interface{}
			for _, header := range headers {
				headerInterfaces = append(headerInterfaces, header)
			}

			httpEntry["headers"] = schema.NewSet(
				schema.HashResource(HTTPHealthCheckHeaderSchema()),
				headerInterfaces,
			)

			r["http"] = schema.NewSet(
				schema.HashResource(HTTPHealthCheckSchema()),
				[]interface{}{httpEntry},
			)
		}
		result[i] = r
	}

	return result
}

func expandVolumes(config []interface{}) []koyeb.DeploymentVolume {
	volumes := make([]koyeb.DeploymentVolume, 0, len(config))

	for _, rawVolume := range config {
		volume := rawVolume.(map[string]interface{})

		v := koyeb.DeploymentVolume{
			Id:           toOpt(volume["id"].(string)),
			Path:         toOpt(volume["path"].(string)),
			ReplicaIndex: toOpt(int64(volume["replica_index"].(int))),
		}

		rawScopes := volume["scopes"].([]interface{})
		scopes := make([]string, len(rawScopes))
		for i, v := range rawScopes {
			scopes[i] = v.(string)
		}
		v.Scopes = scopes

		volumes = append(volumes, v)
	}

	return volumes
}

func flattenVolumes(volumes *[]koyeb.DeploymentVolume) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*volumes))

	for i, volume := range *volumes {
		r := make(map[string]interface{})

		r["id"] = volume.GetId()
		r["path"] = volume.GetPath()
		r["replica_index"] = volume.GetReplicaIndex()
		// r["scopes"] = volume.GetScopes()

		result[i] = r
	}

	return result
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
	if check, ok := deployment.GetHealthChecksOk(); ok {
		r["health_checks"] = flattenHealthChecks(toOpt(check))
	}
	r["routes"] = flattenRoutes(toOpt(deployment.GetRoutes()))
	r["instance_types"] = flattenInstanceTypes(toOpt(deployment.GetInstanceTypes()))
	r["scalings"] = flattenScalings(toOpt(deployment.GetScalings()))
	r["regions"] = flattenRegions(&deployment.Regions)
	r["volumes"] = flattenVolumes(&deployment.Volumes)

	result = append(result, r)

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
	mapper := idmapper.NewMapper(context.Background(), client)
	appMapper := mapper.App()
	var appId string

	if d.Get("app_name").(string) != "" {
		id, err := appMapper.ResolveID(d.Get("app_name").(string))
		if err != nil {
			return diag.Errorf("Error creating service: %s", err)
		}

		appId = id
	}

	definition := expandDeploymentDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{}))

	res, resp, err := client.ServicesApi.CreateService(context.Background()).Service(koyeb.CreateService{
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
	mapper := idmapper.NewMapper(context.Background(), client)
	serviceMapper := mapper.Service()
	var serviceId string

	if d.Id() != "" {
		id, err := serviceMapper.ResolveID(d.Id())
		if err != nil {
			return diag.Errorf("Error retrieving service: %s", err)
		}

		serviceId = id
	}

	serviceRes, resp, err := client.ServicesApi.GetService(context.Background(), serviceId).Execute()
	if err != nil {
		// If the service is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving service: %s (%v %v)", err, resp, serviceRes)
	}

	// deploymentRes, resp, err := client.DeploymentsApi.GetDeployment(context.Background(), *serviceRes.Service.LatestDeploymentId).Execute()
	// if err != nil {
	// 	return diag.Errorf("Error retrieving service latest deployment: %s (%v %v", err, resp, serviceRes)
	// }

	setServiceAttribute(d, serviceRes.Service)

	return nil
}

func resourceKoyebServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	definition := expandDeploymentDefinition(d.Get("definition").([]interface{})[0].(map[string]interface{}))
	res, resp, err := client.ServicesApi.UpdateService(context.Background(), d.Id()).Service(koyeb.UpdateService{
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

	res, resp, err := client.ServicesApi.DeleteService(context.Background(), d.Id()).Execute()
	if err != nil {
		return diag.Errorf("Error deleting service: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
