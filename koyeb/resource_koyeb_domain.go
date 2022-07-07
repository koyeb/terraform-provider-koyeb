package koyeb

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func domainSchema() map[string]*schema.Schema {
	domain := map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Description:  "The domain name",
			ForceNew:     true,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The id of the domain",
		},
		"version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The version of the domain",
		},
		"deployment_group": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The deployment group assigned to the domain",
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization id owning the domain",
		},
		"app_name": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "The app name the domain is assigned to",
			ValidateFunc: validation.StringLenBetween(3, 23),
		},
		"type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The domain type",
		},
		"intended_cname": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The CNAME record to point the domain to",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The status of the domain",
		},
		"messages": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The status messages of the domain",
		},
		"verified_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Optional:    true,
			Description: "The date and time of when the domain was last verified",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the domain was last updated",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the domain was created",
		},
	}

	return domain
}

func flattenDomains(domains *[]koyeb.Domain, appName string) []map[string]interface{} {
	result := make([]map[string]interface{}, len(*domains))

	for i, domain := range *domains {
		r := make(map[string]interface{})

		r["name"] = domain.GetName()
		r["id"] = domain.GetId()
		r["type"] = domain.GetType()
		r["status"] = domain.GetStatus()
		r["version"] = domain.GetVersion()
		r["deployment_group"] = domain.GetDeploymentGroup()
		r["organization_id"] = domain.GetOrganizationId()
		r["created_at"] = domain.GetCreatedAt().UTC().String()
		r["updated_at"] = domain.GetUpdatedAt().UTC().String()
		r["app_name"] = appName
		if messages, ok := domain.GetMessagesOk(); ok && len(domain.GetMessages()) > 0 {
			r["messages"] = strings.Join(*messages, " ")
		}

		if verifiedAt, ok := domain.GetVerifiedAtOk(); ok {
			r["verified_at"] = verifiedAt
		}

		if intendedCname, ok := domain.GetIntendedCnameOk(); ok {
			r["intended_cname"] = intendedCname
		}

		result[i] = r
	}
	return result
}

func resourceKoyebDomain() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Domain resource in the Koyeb Terraform provider.",

		CreateContext: resourceKoyebDomainCreate,
		ReadContext:   resourceKoyebDomainRead,
		UpdateContext: resourceKoyebDomainUpdate,
		DeleteContext: resourceKoyebDomainDelete,

		Schema: domainSchema(),
	}
}

func setDomainAttribute(
	d *schema.ResourceData,
	domain *koyeb.Domain,
	appName string,
) error {
	d.SetId(domain.GetId())
	d.Set("id", domain.GetId())
	d.Set("name", domain.GetName())
	d.Set("version", domain.GetVersion())
	d.Set("status", domain.GetStatus())
	d.Set("type", domain.GetType())
	d.Set("messages", strings.Join(domain.GetMessages(), " "))
	d.Set("deployment_group", domain.GetDeploymentGroup())
	d.Set("organization_id", domain.GetOrganizationId())
	d.Set("intended_cname", domain.GetIntendedCname())
	d.Set("verified_at", domain.GetVerifiedAt().UTC().String())
	d.Set("created_at", domain.GetCreatedAt().UTC().String())
	d.Set("updated_at", domain.GetUpdatedAt().UTC().String())
	d.Set("app_name", appName)
	return nil
}

func resourceKoyebDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	mapper := idmapper.NewMapper(ctx, client)
	appMapper := mapper.App()
	var appId string

	if d.Get("app_name").(string) != "" {
		id, err := appMapper.ResolveID(d.Get("app_name").(string))

		if err != nil {
			return diag.Errorf("Error creating domain: %s", err)
		}

		appId = id
	}

	res, resp, err := client.DomainsApi.CreateDomain(ctx).Body(koyeb.CreateDomain{
		Name:  Ptr(d.Get("name").(string)),
		AppId: &appId,
		Type:  Ptr(koyeb.DOMAINTYPE_CUSTOM),
	}).Execute()
	if err != nil {
		return diag.Errorf("Error creating domain: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.Domain.Id)
	log.Printf("[INFO] Created domain name: %s", *res.Domain.Name)

	return resourceKoyebDomainRead(ctx, d, meta)
}

func resourceKoyebDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	appName := ""

	res, resp, err := client.DomainsApi.GetDomain(ctx, d.Id()).Execute()
	if err != nil {
		// If the domain is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving domain: %s (%v %v", err, resp, res)
	}

	if *res.Domain.AppId != "" {
		res, resp, err := client.AppsApi.GetApp(ctx, *res.Domain.AppId).Execute()
		if err != nil {
			return diag.Errorf("Error retrieving app assigned to domain: %s (%v %v)", err, resp, res)
		}

		appName = *res.App.Name
	}

	setDomainAttribute(d, res.Domain, appName)

	return nil
}

func resourceKoyebDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	mapper := idmapper.NewMapper(ctx, client)
	appMapper := mapper.App()
	var appId string

	if d.Get("app_name").(string) != "" {
		id, err := appMapper.ResolveID(d.Get("app_name").(string))

		if err != nil {
			return diag.Errorf("Error creating domain: %s", err)
		}

		appId = id
	}

	res, resp, err := client.DomainsApi.UpdateDomain(ctx, d.Id()).Body(koyeb.UpdateDomain{AppId: &appId}).Execute()

	if err != nil {
		return diag.Errorf("Error retrieving domain: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Updated domain name: %s", *res.Domain.Name)
	return resourceKoyebDomainRead(ctx, d, meta)
}

func resourceKoyebDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.DomainsApi.DeleteDomain(ctx, d.Id()).Execute()

	if err != nil {
		return diag.Errorf("Error deleting domain: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
