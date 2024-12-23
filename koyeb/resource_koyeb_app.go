package koyeb

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func appSchema() map[string]*schema.Schema {
	app := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The app ID",
		},
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			Description:  "The app name",
			ValidateFunc: validation.StringLenBetween(3, 23),
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization ID owning the app",
		},
		"domains": {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: domainSchema(),
			},
			Computed:    true,
			Description: "The app domains",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the app was last updated",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the app was created",
		},
	}

	return app
}

func resourceKoyebApp() *schema.Resource {
	return &schema.Resource{
		Description: "App resource in the Koyeb Terraform provider.",

		CreateContext: resourceKoyebAppCreate,
		ReadContext:   resourceKoyebAppRead,
		DeleteContext: resourceKoyebAppDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: appSchema(),
	}
}

func setAppAttribute(d *schema.ResourceData, app koyeb.App) error {
	d.SetId(app.GetId())
	d.Set("name", app.GetName())
	d.Set("organization_id", app.GetOrganizationId())
	d.Set("domains", flattenDomains(&app.Domains, app.GetName()))
	d.Set("updated_at", app.GetUpdatedAt().UTC().String())
	d.Set("created_at", app.GetCreatedAt().UTC().String())

	return nil
}

func resourceKoyebAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.AppsApi.CreateApp(context.Background()).App(koyeb.CreateApp{
		Name: toOpt(d.Get("name").(string)),
	}).Execute()

	if err != nil {
		return diag.Errorf("Error creating app: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.App.Id)
	log.Printf("[INFO] Created app name: %s", *res.App.Name)

	return resourceKoyebAppRead(ctx, d, meta)
}

func resourceKoyebAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	mapper := idmapper.NewMapper(context.Background(), client)
	appMapper := mapper.App()
	var appId string

	if d.Id() != "" {
		id, err := appMapper.ResolveID(d.Id())

		if err != nil {
			return diag.Errorf("Error retrieving app: %s", err)
		}

		appId = id
	}

	res, resp, err := client.AppsApi.GetApp(context.Background(), appId).Execute()
	if err != nil {
		// If the app is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving app: %s (%v %v)", err, resp, res)
	}

	setAppAttribute(d, *res.App)

	return nil
}

func resourceKoyebAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.AppsApi.DeleteApp(context.Background(), d.Id()).Execute()

	if err != nil {
		return diag.Errorf("Error deleting app: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
