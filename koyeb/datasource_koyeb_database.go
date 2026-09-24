package koyeb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func databaseDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The database service ID",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The database name; the app and the service both carry this name",
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization ID owning the database",
		},
		"app_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The app ID the database service is assigned to",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the database service",
		},
		"version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The version of the database service",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the database was last updated",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the database was created",
		},
	}
}

func dataSourceKoyebDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "Database data source in the Koyeb Terraform provider.",
		ReadContext: dataSourceKoyebDatabaseRead,
		Schema:      databaseDataSourceSchema(),
	}
}

func dataSourceKoyebDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	mapper := idmapper.NewMapper(context.Background(), client)
	databaseMapper := mapper.Database()

	// A database created by the koyeb_database resource (or the CLI
	// `koyeb db create NAME`) lives in an app and a service both named
	// after the database, so the mapper key is NAME/NAME.
	name := d.Get("name").(string)
	id, err := databaseMapper.ResolveID(name + "/" + name)
	if err != nil {
		return diag.Errorf("Error retrieving database: %s", err)
	}

	d.SetId(id)

	return resourceKoyebDatabaseRead(ctx, d, meta)
}
