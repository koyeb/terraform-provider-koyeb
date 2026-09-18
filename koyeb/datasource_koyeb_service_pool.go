package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func dataSourceKoyebServicePool() *schema.Resource {
	s := servicePoolSchema()
	// The datasource only looks pools up: the attributes the resource
	// owns are read-only here.
	s["size"].Required = false
	s["size"].Computed = true
	s["definition"].Required = false
	s["definition"].Computed = true

	return &schema.Resource{
		ReadContext: dataSourceKoyebServicePoolRead,
		Schema:      s,
	}
}

func dataSourceKoyebServicePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	name := d.Get("name").(string)

	res, resp, err := client.ServicePoolsApi.ListServicePools(context.Background()).Name(name).Execute()
	if err != nil {
		return diag.Errorf("Error retrieving service pool: %s (%v %v)", err, resp, res)
	}

	for _, pool := range res.GetServicePools() {
		if pool.GetName() == name {
			d.SetId(pool.GetId())
			return resourceKoyebServicePoolRead(ctx, d, meta)
		}
	}

	return diag.Errorf("No service pool found with name %s", name)
}
