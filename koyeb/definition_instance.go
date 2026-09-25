package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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
