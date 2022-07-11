package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func dataSourceKoyebDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKoyebDomainRead,
		Schema:      domainSchema(),
	}
}

func dataSourceKoyebDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(context.Background(), client)
	domainMapper := mapper.Domain()

	id, err := domainMapper.ResolveID(d.Get("name").(string))

	if err != nil {
		return diag.Errorf("Error retrieving domain: %s", err)
	}

	d.SetId(id)

	return resourceKoyebDomainRead(ctx, d, meta)
}
