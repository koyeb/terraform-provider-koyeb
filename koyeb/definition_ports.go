package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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

func proxyPortSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "The port exposed by the proxy port",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The protocol used by the proxy port",
				ValidateFunc: validation.StringInSlice([]string{
					"tcp",
				}, false),
			},
		},
	}
}

func flattenProxyPorts(proxyPorts *[]koyeb.DeploymentProxyPort) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*proxyPorts))

	for i, proxyPort := range *proxyPorts {
		r := make(map[string]interface{})

		r["port"] = int(proxyPort.GetPort())
		r["protocol"] = string(proxyPort.GetProtocol())

		result[i] = r
	}

	return result
}

func expandProxyPorts(config []interface{}) []koyeb.DeploymentProxyPort {
	proxyPorts := make([]koyeb.DeploymentProxyPort, 0, len(config))

	for _, rawProxyPort := range config {
		proxyPort := rawProxyPort.(map[string]interface{})

		expanded := koyeb.DeploymentProxyPort{
			Port: toOpt(int64(proxyPort["port"].(int))),
		}
		if protocol, ok := proxyPort["protocol"].(string); ok && protocol != "" {
			expanded.Protocol = toOpt(koyeb.ProxyPortProtocol(protocol))
		}

		proxyPorts = append(proxyPorts, expanded)
	}

	return proxyPorts
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
