package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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
