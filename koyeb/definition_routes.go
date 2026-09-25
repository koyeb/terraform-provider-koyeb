package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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
			"security_policies": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "The security policies applied to the route",
				Elem:        securityPoliciesSchema(),
				Set:         schema.HashResource(securityPoliciesSchema()),
			},
		},
	}
}

func securityPoliciesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"basic_auths": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The basic auth credentials protecting the route",
				Elem:        basicAuthPolicySchema(),
				Set:         schema.HashResource(basicAuthPolicySchema()),
			},
			"api_keys": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The API keys protecting the route",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
		},
	}
}

func basicAuthPolicySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The basic auth username",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The basic auth password",
			},
		},
	}
}

func expandSecurityPolicies(config []interface{}) *koyeb.SecurityPolicies {
	if len(config) == 0 {
		return nil
	}

	rawSecurityPolicies := config[0].(map[string]interface{})
	securityPolicies := &koyeb.SecurityPolicies{}

	for _, rawBasicAuth := range rawSecurityPolicies["basic_auths"].(*schema.Set).List() {
		basicAuth := rawBasicAuth.(map[string]interface{})
		securityPolicies.BasicAuths = append(securityPolicies.BasicAuths, koyeb.BasicAuthPolicy{
			Username: toOpt(basicAuth["username"].(string)),
			Password: toOpt(basicAuth["password"].(string)),
		})
	}

	for _, apiKey := range rawSecurityPolicies["api_keys"].(*schema.Set).List() {
		securityPolicies.ApiKeys = append(securityPolicies.ApiKeys, apiKey.(string))
	}

	return securityPolicies
}

func expandRoutes(config []interface{}) []koyeb.DeploymentRoute {
	routes := make([]koyeb.DeploymentRoute, 0, len(config))

	for _, rawRoute := range config {
		route := rawRoute.(map[string]interface{})

		r := koyeb.DeploymentRoute{
			Port: toOpt(int64(route["port"].(int))),
			Path: toOpt(route["path"].(string)),
		}

		securityPolicies := route["security_policies"].(*schema.Set).List()
		if len(securityPolicies) > 0 {
			r.SecurityPolicies = expandSecurityPolicies(securityPolicies)
		}

		routes = append(routes, r)
	}

	return routes
}

func flattenSecurityPolicies(securityPolicies *koyeb.SecurityPolicies) *schema.Set {
	basicAuths := make([]interface{}, 0, len(securityPolicies.BasicAuths))
	for _, basicAuth := range securityPolicies.BasicAuths {
		basicAuths = append(basicAuths, map[string]interface{}{
			"username": basicAuth.GetUsername(),
			"password": basicAuth.GetPassword(),
		})
	}

	apiKeys := make([]interface{}, 0, len(securityPolicies.ApiKeys))
	for _, apiKey := range securityPolicies.ApiKeys {
		apiKeys = append(apiKeys, apiKey)
	}

	return schema.NewSet(
		schema.HashResource(securityPoliciesSchema()),
		[]interface{}{map[string]interface{}{
			"basic_auths": schema.NewSet(
				schema.HashResource(basicAuthPolicySchema()),
				basicAuths,
			),
			"api_keys": schema.NewSet(schema.HashString, apiKeys),
		}},
	)
}

func flattenRoutes(routes *[]koyeb.DeploymentRoute) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*routes))

	for i, route := range *routes {
		r := make(map[string]interface{})

		r["port"] = route.GetPort()
		r["path"] = route.GetPath()
		if securityPolicies, ok := route.GetSecurityPoliciesOk(); ok {
			r["security_policies"] = flattenSecurityPolicies(securityPolicies)
		}

		result[i] = r
	}

	return result
}
