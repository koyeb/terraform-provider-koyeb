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
	// MinItems/MaxItems are only valid on configurable attributes.
	s["definition"].MinItems = 0
	s["definition"].MaxItems = 0

	return &schema.Resource{
		ReadContext: dataSourceKoyebServicePoolRead,
		Schema:      s,
	}
}

func dataSourceKoyebServicePoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	name := d.Get("name").(string)

	id, found, err := findPoolByName(ctx, client, name)
	if err != nil {
		return diag.Errorf("Error retrieving service pool: %s", err)
	}
	if !found {
		return diag.Errorf("No service pool found with name %s", name)
	}
	d.SetId(id)
	return resourceKoyebServicePoolRead(ctx, d, meta)
}
