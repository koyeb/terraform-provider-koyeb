package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func dataSourceKoyebApp() *schema.Resource {
	// The app schema is shared with the resource; on a data source the
	// lifecycle flag can only ever be read, never configured.
	s := appSchema()
	s["delete_when_empty"].Optional = false

	return &schema.Resource{
		ReadContext: dataSourceKoyebAppRead,
		Schema:      s,
	}
}

func dataSourceKoyebAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(context.Background(), client)
	appMapper := mapper.App()

	id, err := appMapper.ResolveID(d.Get("name").(string))

	if err != nil {
		return diag.Errorf("Error retrieving app: %s", err)
	}

	d.SetId(id)

	return resourceKoyebAppRead(ctx, d, meta)
}
