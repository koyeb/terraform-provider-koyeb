package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func configFileSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The absolute path where the config file is mounted in the service",
			},
			"content": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The content of the config file",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The file permissions, e.g. 0644",
				ValidateFunc: validation.StringMatch(configFilePermissionsRegex,
					"must be an octal permission string like 0644"),
			},
		},
	}
}

func flattenConfigFiles(configFiles *[]koyeb.ConfigFile) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*configFiles))

	for i, configFile := range *configFiles {
		r := make(map[string]interface{})

		r["path"] = configFile.GetPath()
		r["content"] = configFile.GetContent()
		r["permissions"] = configFile.GetPermissions()

		result[i] = r
	}

	return result
}

func expandConfigFiles(config []interface{}) []koyeb.ConfigFile {
	configFiles := make([]koyeb.ConfigFile, 0, len(config))

	for _, rawConfigFile := range config {
		configFile := rawConfigFile.(map[string]interface{})

		configFiles = append(configFiles, koyeb.ConfigFile{
			Path:        toOpt(configFile["path"].(string)),
			Content:     toOpt(configFile["content"].(string)),
			Permissions: toOpt(configFile["permissions"].(string)),
		})
	}

	return configFiles
}
