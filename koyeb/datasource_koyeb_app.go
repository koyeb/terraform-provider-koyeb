package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func dataSourceKoyebApp() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKoyebAppRead,
		Schema:      appSchema(),
	}
}

func dataSourceKoyebAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(ctx, client)
	appMapper := mapper.App()

	id, err := appMapper.ResolveID(d.Get("name").(string))

	if err != nil {
		return diag.Errorf("Error retrieving app: %s %v", err, d.Get("name"))
	}

	d.SetId(id)

	return resourceKoyebAppRead(ctx, d, meta)
}
