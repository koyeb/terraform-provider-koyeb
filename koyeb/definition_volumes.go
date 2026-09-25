package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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

func expandVolumes(config []interface{}) []koyeb.DeploymentVolume {
	volumes := make([]koyeb.DeploymentVolume, 0, len(config))

	for _, rawVolume := range config {
		volume := rawVolume.(map[string]interface{})

		v := koyeb.DeploymentVolume{
			Id:           toOpt(volume["id"].(string)),
			Path:         toOpt(volume["path"].(string)),
			ReplicaIndex: toOpt(int64(volume["replica_index"].(int))),
		}

		// The schema key is `scope` (singular); the API field is `scopes`.
		rawScopes := volume["scope"].([]interface{})
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
		r["scope"] = volume.GetScopes()

		result[i] = r
	}

	return result
}
