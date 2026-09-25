package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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
