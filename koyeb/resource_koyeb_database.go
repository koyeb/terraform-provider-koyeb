package koyeb

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func databaseSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The database service ID",
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The database name; an app with the same name is created if it does not exist",
		},
		"pg_version": {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     16,
			Description: "The PostgreSQL version",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "was",
			Description: "The region where the database is deployed",
		},
		"instance_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "free",
			Description: "The instance type of the database (free, small, medium or large)",
		},
		"db_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "koyebdb",
			Description: "The name of the database created in the PostgreSQL instance",
		},
		"db_owner": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "koyeb-adm",
			Description: "The role owning the database",
		},
		"role_secret": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The name of the managed secret holding the database role password",
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

func resourceKoyebDatabase() *schema.Resource {
	return &schema.Resource{
		Description: "Database resource in the Koyeb Terraform provider. A database is deployed as a service of type DATABASE inside an app named after the database; deleting the database deletes the service but leaves the app in place.",

		CreateContext: resourceKoyebDatabaseCreate,
		ReadContext:   resourceKoyebDatabaseRead,
		UpdateContext: resourceKoyebDatabaseUpdate,
		DeleteContext: resourceKoyebDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: databaseSchema(),
	}
}

func setDatabaseAttribute(d *schema.ResourceData, service koyeb.Service) error {
	d.SetId(service.GetId())
	d.Set("name", service.GetName())
	d.Set("app_id", service.GetAppId())
	d.Set("organization_id", service.GetOrganizationId())
	d.Set("status", service.GetStatus())
	d.Set("version", service.GetVersion())
	d.Set("updated_at", service.GetUpdatedAt().UTC().String())
	d.Set("created_at", service.GetCreatedAt().UTC().String())

	return nil
}

func resourceKoyebDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	mapper := idmapper.NewMapper(context.Background(), client)

	_, _, err := client.AppsApi.CreateApp(context.Background()).App(koyeb.CreateApp{
		Name: toOpt(d.Get("name").(string)),
	}).Execute()
	if err != nil && !isAppNameAlreadyExistsError(err) {
		return diag.Errorf("Error creating the database app: %s", err)
	}

	appID, err := mapper.App().ResolveID(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("Error resolving the database app: %s", err)
	}

	roleSecret, err := uuid.GenerateUUID()
	if err != nil {
		return diag.Errorf("Error generating the database role secret name: %s", err)
	}
	d.Set("role_secret", roleSecret)

	definition := buildDatabaseDefinition(
		d.Get("name").(string),
		d.Get("pg_version").(int),
		d.Get("region").(string),
		d.Get("instance_type").(string),
		d.Get("db_name").(string),
		d.Get("db_owner").(string),
		roleSecret,
	)

	res, resp, err := client.ServicesApi.CreateService(context.Background()).Service(koyeb.CreateService{
		AppId:      toOpt(appID),
		Definition: definition,
	}).Execute()
	if err != nil {
		return diag.Errorf("Error creating database: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.Service.Id)
	log.Printf("[INFO] Created database name: %s", *res.Service.Name)

	return resourceKoyebDatabaseRead(ctx, d, meta)
}

func resourceKoyebDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	mapper := idmapper.NewMapper(context.Background(), client)
	serviceMapper := mapper.Service()

	databaseId, err := serviceMapper.ResolveID(d.Id())
	if err != nil {
		return diag.Errorf("Error retrieving database: %s", err)
	}

	res, resp, err := client.ServicesApi.GetService(context.Background(), databaseId).Execute()
	if err != nil {
		// If the database is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving database: %s (%v %v)", err, resp, res)
	}

	setDatabaseAttribute(d, *res.Service)

	return nil
}

func resourceKoyebDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleSecret := d.Get("role_secret").(string)
	if roleSecret == "" {
		return diag.Errorf("role_secret is unknown: update koyeb_database %s after a terraform import would silently redefine the database role; recreate the resource or repair its state", d.Id())
	}

	client := meta.(*koyeb.APIClient)

	definition := buildDatabaseDefinition(
		d.Get("name").(string),
		d.Get("pg_version").(int),
		d.Get("region").(string),
		d.Get("instance_type").(string),
		d.Get("db_name").(string),
		d.Get("db_owner").(string),
		roleSecret,
	)

	res, resp, err := client.ServicesApi.UpdateService(context.Background(), d.Id()).Service(koyeb.UpdateService{
		Definition: definition,
	}).Execute()
	if err != nil {
		return diag.Errorf("Error updating database: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Updated database name: %s", *res.Service.Name)

	return resourceKoyebDatabaseRead(ctx, d, meta)
}

func resourceKoyebDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.ServicesApi.DeleteService(context.Background(), d.Id()).Execute()
	if err != nil {
		return diag.Errorf("Error deleting database: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}

func buildDatabaseDefinition(name string, pgVersion int, region, instanceType, dbName, dbOwner, roleSecret string) *koyeb.DeploymentDefinition {
	return &koyeb.DeploymentDefinition{
		Name: toOpt(name),
		Type: toOpt(koyeb.DEPLOYMENTDEFINITIONTYPE_DATABASE),
		Database: &koyeb.DatabaseSource{
			NeonPostgres: &koyeb.NeonPostgresDatabase{
				PgVersion:    toOpt(int64(pgVersion)),
				Region:       toOpt(region),
				InstanceType: toOpt(instanceType),
				Databases: []koyeb.NeonPostgresDatabaseNeonDatabase{
					{
						Name:  toOpt(dbName),
						Owner: toOpt(dbOwner),
					},
				},
				Roles: []koyeb.NeonPostgresDatabaseNeonRole{
					{
						Name:   toOpt(dbOwner),
						Secret: toOpt(roleSecret),
					},
				},
			},
		},
	}
}

// isAppNameAlreadyExistsError reports whether the CreateApp call failed
// because an app with the same name already exists. The API does not return
// a dedicated error code for this case, the field error is all we get.
func isAppNameAlreadyExistsError(err error) bool {
	var openAPIError *koyeb.GenericOpenAPIError
	if !errors.As(err, &openAPIError) {
		return false
	}

	errorWithFields, ok := openAPIError.Model().(koyeb.ErrorWithFields)
	if !ok {
		return false
	}

	fields := errorWithFields.GetFields()
	return len(fields) == 1 && fields[0].GetField() == "name" && strings.Contains(strings.ToLower(fields[0].GetDescription()), "already exists")
}
