package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func dataSourceKoyebSecret() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKoyebSecretRead,
		Schema:      secretSchema(),
	}
}

func dataSourceKoyebSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(context.Background(), client)
	SecretMapper := mapper.Secret()

	id, err := SecretMapper.ResolveID(d.Get("name").(string))

	if err != nil {
		return diag.Errorf("Error retrieving secret: %s", err)
	}

	d.SetId(id)

	return resourceKoyebSecretRead(ctx, d, meta)
}
