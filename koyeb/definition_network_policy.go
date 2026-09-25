package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func networkPolicySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"egress": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "The egress policy of the service",
				Elem:        egressPolicySchema(),
				Set:         schema.HashResource(egressPolicySchema()),
			},
			"mesh": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "Which services can reach this service through the mesh",
				Elem:        meshPolicySchema(),
				Set:         schema.HashResource(meshPolicySchema()),
			},
		},
	}
}

func egressPolicySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The egress mode: EGRESS_POLICY_MODE_DEFAULT or EGRESS_POLICY_MODE_DENY_ALL",
				ValidateFunc: validation.StringInSlice([]string{
					"EGRESS_POLICY_MODE_DEFAULT",
					"EGRESS_POLICY_MODE_DENY_ALL",
				}, false),
			},
			"allow_list": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The allowed destinations when the egress mode is EGRESS_POLICY_MODE_DENY_ALL, as IPv4 or IPv6 CIDRs",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
		},
	}
}

func meshPolicySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"scope": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The mesh scope: MESH_SCOPE_UNSPECIFIED, MESH_SCOPE_ORGANIZATION, MESH_SCOPE_WORKSPACE, MESH_SCOPE_APP or MESH_SCOPE_CUSTOM",
				ValidateFunc: validation.StringInSlice([]string{
					"MESH_SCOPE_UNSPECIFIED",
					"MESH_SCOPE_ORGANIZATION",
					"MESH_SCOPE_WORKSPACE",
					"MESH_SCOPE_APP",
					"MESH_SCOPE_CUSTOM",
				}, false),
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A custom mesh name, required when the scope is MESH_SCOPE_CUSTOM",
			},
		},
	}
}

func expandNetworkPolicy(config []interface{}) *koyeb.NetworkPolicy {
	if len(config) == 0 {
		return nil
	}

	rawNetworkPolicy := config[0].(map[string]interface{})
	networkPolicy := &koyeb.NetworkPolicy{}

	egress := rawNetworkPolicy["egress"].(*schema.Set).List()
	if len(egress) > 0 {
		rawEgress := egress[0].(map[string]interface{})
		egressPolicy := &koyeb.EgressPolicy{}
		if mode, ok := rawEgress["mode"].(string); ok && mode != "" {
			egressPolicy.Mode = toOpt(koyeb.EgressPolicyMode(mode))
		}
		for _, destination := range rawEgress["allow_list"].(*schema.Set).List() {
			egressPolicy.AllowList = append(egressPolicy.AllowList, koyeb.NetworkPolicyDestination{
				Cidr: toOpt(destination.(string)),
			})
		}
		networkPolicy.Egress = egressPolicy
	}

	mesh := rawNetworkPolicy["mesh"].(*schema.Set).List()
	if len(mesh) > 0 {
		rawMesh := mesh[0].(map[string]interface{})
		meshPolicy := &koyeb.Mesh{}
		if scope, ok := rawMesh["scope"].(string); ok && scope != "" {
			meshPolicy.Scope = toOpt(koyeb.MeshScope(scope))
		}
		if name, ok := rawMesh["name"].(string); ok && name != "" {
			meshPolicy.Name = toOpt(name)
		}
		networkPolicy.Mesh = meshPolicy
	}

	return networkPolicy
}

func flattenNetworkPolicy(networkPolicy *koyeb.NetworkPolicy) []interface{} {
	if networkPolicy.Egress == nil && networkPolicy.Mesh == nil {
		return []interface{}{}
	}

	result := make(map[string]interface{})

	if networkPolicy.Egress != nil {
		egress := map[string]interface{}{}
		if mode, ok := networkPolicy.Egress.GetModeOk(); ok {
			egress["mode"] = string(*mode)
		}
		allowList := make([]interface{}, 0, len(networkPolicy.Egress.AllowList))
		for _, destination := range networkPolicy.Egress.AllowList {
			allowList = append(allowList, destination.GetCidr())
		}
		egress["allow_list"] = schema.NewSet(schema.HashString, allowList)

		result["egress"] = schema.NewSet(
			schema.HashResource(egressPolicySchema()),
			[]interface{}{egress},
		)
	}

	if networkPolicy.Mesh != nil {
		mesh := map[string]interface{}{
			"name": networkPolicy.Mesh.GetName(),
		}
		if scope, ok := networkPolicy.Mesh.GetScopeOk(); ok {
			mesh["scope"] = string(*scope)
		}

		result["mesh"] = schema.NewSet(
			schema.HashResource(meshPolicySchema()),
			[]interface{}{mesh},
		)
	}

	return []interface{}{result}
}
