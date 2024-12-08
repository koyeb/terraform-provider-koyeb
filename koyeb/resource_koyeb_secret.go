package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

const (
	defaultSecretType  = "SIMPLE"
	secretTypeSimple   = "SIMPLE"
	secretTypeRegistry = "REGISTRY"
)

var registryTypes = []string{
	"docker_hub_registry",
	"github_registry",
	"gitlab_registry",
	"digital_ocean_container_registry",
	"private_registry",
	"azure_container_registry",
}

func generateConflictRules(current string, others []string) []string {
	var conflicts []string
	for _, item := range others {
		if item != current {
			conflicts = append(conflicts, item)
		}
	}
	return conflicts
}

func createRegistrySchema(additionalFields map[string]*schema.Schema) *schema.Resource {
	baseFields := map[string]*schema.Schema{
		"username": {
			Type:        schema.TypeString,
			Description: "The registry username",
			Required:    true,
		},
		"password": {
			Type:        schema.TypeString,
			Description: "The registry password",
			Required:    true,
			Sensitive:   true,
		},
	}

	for key, field := range additionalFields {
		baseFields[key] = field
	}

	return &schema.Resource{Schema: baseFields}
}
func secretSchema() map[string]*schema.Schema {
	registrySchemas := map[string]*schema.Resource{
		"docker_hub_registry":              createRegistrySchema(nil),
		"github_registry":                  createRegistrySchema(nil),
		"gitlab_registry":                  createRegistrySchema(nil),
		"digital_ocean_container_registry": createRegistrySchema(nil),
		"private_registry": createRegistrySchema(map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Description: "The registry URL",
				Required:    true,
			},
		}),
		"azure_container_registry": createRegistrySchema(map[string]*schema.Schema{
			"registry_name": {
				Type:        schema.TypeString,
				Description: "The registry name",
				Required:    true,
			},
		}),
	}

	schemaMap := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The secret ID",
		},
		"name": {
			Type:         schema.TypeString,
			Description:  "The secret name",
			ForceNew:     true,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(2, 64),
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization ID owning the secret",
		},
		"type": {
			Type:         schema.TypeString,
			ForceNew:     true,
			Optional:     true,
			Default:      defaultSecretType,
			Description:  "The secret type",
			ValidateFunc: validation.StringInSlice([]string{secretTypeSimple, secretTypeRegistry}, false),
		},
		"value": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "The secret value",
			Sensitive:     true,
			ConflictsWith: registryTypes,
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the secret was last updated",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the secret was created",
		},
	}

	for key, resource := range registrySchemas {
		schemaMap[key] = &schema.Schema{
			Type:          schema.TypeSet,
			Optional:      true,
			Elem:          resource,
			Description:   fmt.Sprintf("The %s configuration to use", key),
			MaxItems:      1,
			ConflictsWith: generateConflictRules(key, registryTypes),
		}
	}

	return schemaMap
}

func expandRegistry(config []interface{}, registryType string) interface{} {
	if len(config) == 0 {
		return nil
	}

	rawRegistry := config[0].(map[string]interface{})
	log.Printf("Expanding registry: %v", rawRegistry)

	switch registryType {
	case "docker_hub_registry":
		return &koyeb.DockerHubRegistryConfiguration{
			Username: toOpt(rawRegistry["username"].(string)),
			Password: toOpt(rawRegistry["password"].(string)),
		}
	case "github_registry":
		return &koyeb.GitHubRegistryConfiguration{
			Username: toOpt(rawRegistry["username"].(string)),
			Password: toOpt(rawRegistry["password"].(string)),
		}
	case "gitlab_registry":
		return &koyeb.GitLabRegistryConfiguration{
			Username: toOpt(rawRegistry["username"].(string)),
			Password: toOpt(rawRegistry["password"].(string)),
		}
	case "digital_ocean_container_registry":
		return &koyeb.DigitalOceanRegistryConfiguration{
			Username: toOpt(rawRegistry["username"].(string)),
			Password: toOpt(rawRegistry["password"].(string)),
		}
	case "private_registry":
		return &koyeb.PrivateRegistryConfiguration{
			Username: toOpt(rawRegistry["username"].(string)),
			Password: toOpt(rawRegistry["password"].(string)),
			Url:      toOpt(rawRegistry["url"].(string)),
		}
	case "azure_container_registry":
		return &koyeb.AzureContainerRegistryConfiguration{
			Username:     toOpt(rawRegistry["username"].(string)),
			Password:     toOpt(rawRegistry["password"].(string)),
			RegistryName: toOpt(rawRegistry["registry_name"].(string)),
		}

	default:
		return nil
	}
}

func flattenRegistry(data map[string]interface{}, keys ...string) []interface{} {
	flattened := make(map[string]interface{})
	for _, key := range keys {
		if value, ok := data[key]; ok {
			flattened[key] = value
		}
	}
	return []interface{}{flattened}
}

func resourceKoyebSecret() *schema.Resource {
	return &schema.Resource{
		Description: "Secret resource in the Koyeb Terraform provider.",

		CreateContext: resourceKoyebSecretCreate,
		ReadContext:   resourceKoyebSecretRead,
		UpdateContext: resourceKoyebSecretUpdate,
		DeleteContext: resourceKoyebSecretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: secretSchema(),
	}
}

func setSecretAttribute(d *schema.ResourceData, secret koyeb.Secret, secretValue interface{}) error {
	d.SetId(secret.GetId())
	d.Set("name", secret.GetName())
	d.Set("organization_id", secret.GetOrganizationId())
	d.Set("type", secret.GetType())

	if _, ok := secret.GetValueOk(); ok {
		d.Set("value", secretValue)
	}
	if _, ok := secret.GetDockerHubRegistryOk(); ok {
		d.Set("docker_hub_registry", flattenRegistry(secretValue.(map[string]interface{}), "username", "password"))
	}
	if _, ok := secret.GetGithubRegistryOk(); ok {
		d.Set("github_registry", flattenRegistry(secretValue.(map[string]interface{}), "username", "password"))
	}
	if _, ok := secret.GetGitlabRegistryOk(); ok {
		d.Set("gitlab_registry", flattenRegistry(secretValue.(map[string]interface{}), "username", "password"))
	}
	if _, ok := secret.GetDigitalOceanRegistryOk(); ok {
		d.Set("digital_ocean_container_registry", flattenRegistry(secretValue.(map[string]interface{}), "username", "password"))
	}
	if _, ok := secret.GetPrivateRegistryOk(); ok {
		d.Set("private_registry", flattenRegistry(secretValue.(map[string]interface{}), "username", "password", "url"))
	}
	if _, ok := secret.GetAzureContainerRegistryOk(); ok {
		d.Set("azure_container_registry", flattenRegistry(secretValue.(map[string]interface{}), "username", "password", "registry_name"))
	}

	d.Set("updated_at", secret.GetUpdatedAt().UTC().String())
	d.Set("created_at", secret.GetCreatedAt().UTC().String())

	return nil
}

func resourceKoyebSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	secret := koyeb.CreateSecret{
		Name: toOpt(d.Get("name").(string)),
		Type: toOpt(koyeb.SecretType(d.Get("type").(string))),
	}

	if value, ok := d.GetOk("value"); ok {
		secret.Value = toOpt(value.(string))
	}

	if dockerHubRegistry, ok := d.GetOk("docker_hub_registry"); ok {
		secret.DockerHubRegistry = expandRegistry(dockerHubRegistry.(*schema.Set).List(), "docker_hub_registry").(*koyeb.DockerHubRegistryConfiguration)
	}

	if gitHubRegistry, ok := d.GetOk("github_registry"); ok {
		secret.GithubRegistry = expandRegistry(gitHubRegistry.(*schema.Set).List(), "github_registry").(*koyeb.GitHubRegistryConfiguration)
	}

	if doRegistry, ok := d.GetOk("digital_ocean_container_registry"); ok {
		secret.DigitalOceanRegistry = expandRegistry(doRegistry.(*schema.Set).List(), "digital_ocean_container_registry").(*koyeb.DigitalOceanRegistryConfiguration)
	}

	if gitLabRegistry, ok := d.GetOk("gitlab_registry"); ok {
		secret.GitlabRegistry = expandRegistry(gitLabRegistry.(*schema.Set).List(), "gitlab_registry").(*koyeb.GitLabRegistryConfiguration)
	}

	if privateRegistry, ok := d.GetOk("private_registry"); ok {
		secret.PrivateRegistry = expandRegistry(privateRegistry.(*schema.Set).List(), "private_registry").(*koyeb.PrivateRegistryConfiguration)
	}

	if azureContainerRegistry, ok := d.GetOk("azure_container_registry"); ok {
		secret.AzureContainerRegistry = expandRegistry(azureContainerRegistry.(*schema.Set).List(), "azure_container_registry").(*koyeb.AzureContainerRegistryConfiguration)
	}

	res, resp, err := client.SecretsApi.CreateSecret(ctx).Secret(secret).Execute()
	if err != nil {
		return diag.Errorf("Error creating secret: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.Secret.Id)
	log.Printf("[INFO] Created secret name: %s", *res.Secret.Name)

	return resourceKoyebSecretRead(ctx, d, meta)
}

func resourceKoyebSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)
	mapper := idmapper.NewMapper(context.Background(), client)
	secretMapper := mapper.Secret()
	var secretId string

	if d.Id() != "" {
		id, err := secretMapper.ResolveID(d.Id())

		if err != nil {
			return diag.Errorf("Error retrieving secret: %s", err)
		}

		secretId = id
	}

	res, resp, err := client.SecretsApi.GetSecret(context.Background(), secretId).Execute()
	if err != nil {
		// If the Secret is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving secret: %s (%v %v)", err, resp, res)
	}

	body := make(map[string]interface{})
	_, resp, err = client.SecretsApi.RevealSecret(context.Background(), secretId).Body(body).Execute()
	if resp.StatusCode != 200 && err != nil {
		return diag.Errorf("Error retrieving secret value: %s", err)

	}

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return diag.Errorf("Error while reading the response body: %s", err)
	}

	output := map[string]interface{}{}
	if err := json.Unmarshal(buffer, &output); err != nil {
		return diag.Errorf("Error while unmarshalling the response body: %s", err)
	}

	secretValue, ok := output["value"]
	if !ok {
		return diag.Errorf("Error while reading the secret value: %s", err)
	}

	setSecretAttribute(d, *res.Secret, secretValue)

	return nil
}

func resourceKoyebSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	secret := koyeb.Secret{
		Name: toOpt(d.Get("name").(string)),
		Type: toOpt(koyeb.SecretType(d.Get("type").(string))),
	}

	if value, ok := d.GetOk("value"); ok {
		secret.Value = toOpt(value.(string))
	}

	if dockerHubRegistry, ok := d.GetOk("docker_hub_registry"); ok {
		secret.DockerHubRegistry = expandRegistry(dockerHubRegistry.(*schema.Set).List(), "docker_hub_registry").(*koyeb.DockerHubRegistryConfiguration)
	}

	if gitHubRegistry, ok := d.GetOk("github_registry"); ok {
		secret.GithubRegistry = expandRegistry(gitHubRegistry.(*schema.Set).List(), "github_registry").(*koyeb.GitHubRegistryConfiguration)
	}

	if doRegistry, ok := d.GetOk("digital_ocean_container_registry"); ok {
		secret.DigitalOceanRegistry = expandRegistry(doRegistry.(*schema.Set).List(), "digital_ocean_container_registry").(*koyeb.DigitalOceanRegistryConfiguration)
	}

	if gitLabRegistry, ok := d.GetOk("gitlab_registry"); ok {
		secret.GitlabRegistry = expandRegistry(gitLabRegistry.(*schema.Set).List(), "gitlab_registry").(*koyeb.GitLabRegistryConfiguration)
	}

	if privateRegistry, ok := d.GetOk("private_registry"); ok {
		secret.PrivateRegistry = expandRegistry(privateRegistry.(*schema.Set).List(), "private_registry").(*koyeb.PrivateRegistryConfiguration)
	}

	if azureContainerRegistry, ok := d.GetOk("azure_container_registry"); ok {
		secret.AzureContainerRegistry = expandRegistry(azureContainerRegistry.(*schema.Set).List(), "azure_container_registry").(*koyeb.AzureContainerRegistryConfiguration)
	}

	res, resp, err := client.SecretsApi.UpdateSecret(context.Background(), d.Id()).Secret(secret).Execute()

	if err != nil {
		return diag.Errorf("Error updating secret: %s (%v %v)", err, resp, res)
	}

	log.Printf("[INFO] Updated secret name: %s", *res.Secret.Name)

	return resourceKoyebSecretRead(ctx, d, meta)
}

func resourceKoyebSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.SecretsApi.DeleteSecret(context.Background(), d.Id()).Execute()

	if err != nil {
		return diag.Errorf("Error deleting secret: %s (%v %v)", err, resp, res)
	}

	d.SetId("")
	return nil
}
