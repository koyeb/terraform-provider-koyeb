package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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
			"sleep_idle_delay": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The delays in seconds after which a service which received 0 request is put to light sleep and deep sleep",
				Elem:        sleepIdleDelayValueSchema(),
				Set:         schema.HashResource(sleepIdleDelayValueSchema()),
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

func sleepIdleDelayValueSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"light_sleep_value": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Delay in seconds after which a service which received 0 request is put to light sleep",
			},
			"deep_sleep_value": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Delay in seconds after which a service which received 0 request is put to deep sleep",
			},
		},
	}
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

			if target["sleep_idle_delay"] != nil {
				sleepIdleDelay := target["sleep_idle_delay"].(*schema.Set).List()
				for _, rawSleepIdleDelay := range sleepIdleDelay {
					sleepIdleDelay := rawSleepIdleDelay.(map[string]interface{})
					sleepIdleDelayTarget := koyeb.DeploymentScalingTargetSleepIdleDelay{}
					if light, ok := sleepIdleDelay["light_sleep_value"]; ok {
						sleepIdleDelayTarget.LightSleepValue = toOpt(int64(light.(int)))
					}
					if deep, ok := sleepIdleDelay["deep_sleep_value"]; ok {
						sleepIdleDelayTarget.DeepSleepValue = toOpt(int64(deep.(int)))
					}
					s.Targets = append(s.Targets, koyeb.DeploymentScalingTarget{
						SleepIdleDelay: &sleepIdleDelayTarget,
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

			if sleepIdleDelay, ok := target.GetSleepIdleDelayOk(); ok {
				targetMap["sleep_idle_delay"] = schema.NewSet(
					schema.HashResource(sleepIdleDelayValueSchema()),
					[]interface{}{
						map[string]interface{}{
							"light_sleep_value": int(sleepIdleDelay.GetLightSleepValue()),
							"deep_sleep_value":  int(sleepIdleDelay.GetDeepSleepValue()),
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
