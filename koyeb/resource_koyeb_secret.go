package koyeb

import (
	"context"
	"encoding/json"
	"io"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
)

func secretSchema() map[string]*schema.Schema {
	secret := map[string]*schema.Schema{
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
			Type:        schema.TypeString,
			ForceNew:    true,
			Optional:    true,
			Default:     "SIMPLE",
			Description: "The secret type",
			ValidateFunc: validation.StringInSlice([]string{
				"SIMPLE",
				"REGISTRY",
			}, false),
		},
		"value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The secret value",
			Sensitive:   true,
			ConflictsWith: []string{
				"docker_hub_registry",
				"github_registry",
				"gitlab_registry",
				"digital_ocean_container_registry",
				"private_registry",
				"azure_container_registry",
			},
		},
		"docker_hub_registry": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        dockerHubRegistrySchema(),
			Description: "The DockerHub registry configuration to use",
			MaxItems:    1,
			ConflictsWith: []string{
				"github_registry",
				"gitlab_registry",
				"digital_ocean_container_registry",
				"private_registry",
				"azure_container_registry",
			},
			Set: schema.HashResource(dockerHubRegistrySchema()),
		},
		"github_registry": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        gitHubRegistrySchema(),
			Description: "The GitHub registry configuration to use",
			MaxItems:    1,
			ConflictsWith: []string{
				"docker_hub_registry",
				"gitlab_registry",
				"digital_ocean_container_registry",
				"private_registry",
				"azure_container_registry",
			},
			Set: schema.HashResource(gitHubRegistrySchema()),
		},
		"gitlab_registry": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        gitLabRegistrySchema(),
			Description: "The GitLab registry configuration to use",
			MaxItems:    1,
			ConflictsWith: []string{
				"docker_hub_registry",
				"github_registry",
				"digital_ocean_container_registry",
				"private_registry",
				"azure_container_registry",
			},
			Set: schema.HashResource(gitLabRegistrySchema()),
		},
		"digital_ocean_container_registry": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        digitalOceanRegistrySchema(),
			Description: "The DigitalOcean registry configuration to use",
			MaxItems:    1,
			ConflictsWith: []string{
				"docker_hub_registry",
				"github_registry",
				"gitlab_registry",
				"private_registry",
				"azure_container_registry",
			},
			Set: schema.HashResource(digitalOceanRegistrySchema()),
		},
		"private_registry": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        privateRegistrySchema(),
			Description: "The DigitalOcean registry configuration to use",
			MaxItems:    1,
			ConflictsWith: []string{
				"docker_hub_registry",
				"github_registry",
				"gitlab_registry",
				"digital_ocean_container_registry",
				"azure_container_registry",
			},
			Set: schema.HashResource(privateRegistrySchema()),
		},
		"azure_container_registry": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        azureRegistrySchema(),
			Description: "The Azure registry configuration to use",
			MaxItems:    1,
			ConflictsWith: []string{
				"docker_hub_registry",
				"github_registry",
				"gitlab_registry",
				"digital_ocean_container_registry",
				"private_registry",
			},
			Set: schema.HashResource(privateRegistrySchema()),
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

	return secret
}

func dockerHubRegistrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func gitHubRegistrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func gitLabRegistrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func digitalOceanRegistrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func privateRegistrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
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
			"url": {
				Type:        schema.TypeString,
				Description: "The registry url",
				Required:    true,
			},
		},
	}
}

func azureRegistrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
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
			"registry_name": {
				Type:        schema.TypeString,
				Description: "The registry name",
				Required:    true,
			},
		},
	}
}

func expandDockerHubRegistry(config []interface{}) *koyeb.DockerHubRegistryConfiguration {
	rawDockerHubRegistry := config[0].(map[string]interface{})

	log.Printf("DockerHubRegistry e: %v", rawDockerHubRegistry)

	dockerHubRegistry := &koyeb.DockerHubRegistryConfiguration{
		Username: toOpt(rawDockerHubRegistry["username"].(string)),
		Password: toOpt(rawDockerHubRegistry["password"].(string)),
	}

	return dockerHubRegistry
}

func flattenDockerHubRegistry(secretValue map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = secretValue["username"]
	r["password"] = secretValue["password"]

	result = append(result, r)

	return result
}

func expandGitHubRegistry(config []interface{}) *koyeb.GitHubRegistryConfiguration {
	rawGitHubRegistry := config[0].(map[string]interface{})

	log.Printf("GitHubRegistry e: %v", rawGitHubRegistry)

	gitHubRegistry := &koyeb.GitHubRegistryConfiguration{
		Username: toOpt(rawGitHubRegistry["username"].(string)),
		Password: toOpt(rawGitHubRegistry["password"].(string)),
	}

	return gitHubRegistry
}

func flattenGitHubRegistry(secretValue map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = secretValue["username"]
	r["password"] = secretValue["password"]

	result = append(result, r)

	return result
}

func expandGitLabRegistry(config []interface{}) *koyeb.GitLabRegistryConfiguration {
	rawGitLabRegistry := config[0].(map[string]interface{})

	gitLabRegistry := &koyeb.GitLabRegistryConfiguration{
		Username: toOpt(rawGitLabRegistry["username"].(string)),
		Password: toOpt(rawGitLabRegistry["password"].(string)),
	}

	return gitLabRegistry
}

func flattenGitLabRegistry(secretValue map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = secretValue["username"]
	r["password"] = secretValue["password"]

	result = append(result, r)

	return result
}

func expandDigitalOceanRegistry(config []interface{}) *koyeb.DigitalOceanRegistryConfiguration {
	rawDigitalOceanRegistry := config[0].(map[string]interface{})

	digitalOceanRegistry := &koyeb.DigitalOceanRegistryConfiguration{
		Username: toOpt(rawDigitalOceanRegistry["username"].(string)),
		Password: toOpt(rawDigitalOceanRegistry["password"].(string)),
	}

	return digitalOceanRegistry
}

func flattenDigitalOceanRegistry(secretValue map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = secretValue["username"]
	r["password"] = secretValue["password"]

	result = append(result, r)

	return result
}

func expandPrivateRegistry(config []interface{}) *koyeb.PrivateRegistryConfiguration {
	rawPrivateRegistry := config[0].(map[string]interface{})

	dockerHubRegistry := &koyeb.PrivateRegistryConfiguration{
		Username: toOpt(rawPrivateRegistry["username"].(string)),
		Password: toOpt(rawPrivateRegistry["password"].(string)),
		Url:      toOpt(rawPrivateRegistry["url"].(string)),
	}

	return dockerHubRegistry
}

func flattenPrivateRegistry(secretValue map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = secretValue["username"]
	r["password"] = secretValue["password"]
	r["url"] = secretValue["url"]

	result = append(result, r)

	return result
}

func expandAzureContainerRegistry(config []interface{}) *koyeb.AzureContainerRegistryConfiguration {
	rawAzureContainerRegistry := config[0].(map[string]interface{})

	azureContainerRegistry := &koyeb.AzureContainerRegistryConfiguration{
		Username:     toOpt(rawAzureContainerRegistry["username"].(string)),
		Password:     toOpt(rawAzureContainerRegistry["password"].(string)),
		RegistryName: toOpt(rawAzureContainerRegistry["registry_name"].(string)),
	}

	return azureContainerRegistry
}

func flattenAzureContainerRegistry(secretValue map[string]interface{}) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = secretValue["username"]
	r["password"] = secretValue["password"]
	r["registry_name"] = secretValue["registry_name"]

	result = append(result, r)

	return result
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
		log.Printf("Setting secret value: %v", secretValue)
		d.Set("value", secretValue)
	}
	if _, ok := secret.GetDockerHubRegistryOk(); ok {
		d.Set("docker_hub_registry", flattenDockerHubRegistry(secretValue.(map[string]interface{})))
	}
	if _, ok := secret.GetGithubRegistryOk(); ok {
		d.Set("github_registry", flattenGitHubRegistry(secretValue.(map[string]interface{})))
	}
	if _, ok := secret.GetGitlabRegistryOk(); ok {
		d.Set("gitlab_registry", flattenGitLabRegistry(secretValue.(map[string]interface{})))
	}
	if _, ok := secret.GetDigitalOceanRegistryOk(); ok {
		d.Set("digital_ocean_container_registry", flattenDigitalOceanRegistry(secretValue.(map[string]interface{})))
	}
	if _, ok := secret.GetPrivateRegistryOk(); ok {
		d.Set("private_registry", flattenPrivateRegistry(secretValue.(map[string]interface{})))
	}
	if _, ok := secret.GetAzureContainerRegistryOk(); ok {
		d.Set("azure_container_registry", flattenAzureContainerRegistry(secretValue.(map[string]interface{})))
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

	if dockerHubRegistry, ok := d.GetOk("docker_hub_registry"); ok && dockerHubRegistry.(*schema.Set).Len() > 0 && dockerHubRegistry.(*schema.Set).List()[0] != nil {
		secret.DockerHubRegistry = expandDockerHubRegistry(d.Get("docker_hub_registry").(*schema.Set).List())
	}

	if gitHubRegistry, ok := d.GetOk("github_registry"); ok && gitHubRegistry.(*schema.Set).Len() > 0 && gitHubRegistry.(*schema.Set).List()[0] != nil {
		secret.GithubRegistry = expandGitHubRegistry(d.Get("github_registry").(*schema.Set).List())
	}

	if doRegistry, ok := d.GetOk("digital_ocean_container_registry"); ok && doRegistry.(*schema.Set).Len() > 0 && doRegistry.(*schema.Set).List()[0] != nil {
		secret.DigitalOceanRegistry = expandDigitalOceanRegistry(d.Get("digital_ocean_container_registry").(*schema.Set).List())
	}

	if gitLabRegistry, ok := d.GetOk("gitlab_registry"); ok && gitLabRegistry.(*schema.Set).Len() > 0 && gitLabRegistry.(*schema.Set).List()[0] != nil {
		secret.GitlabRegistry = expandGitLabRegistry(d.Get("gitlab_registry").(*schema.Set).List())
	}

	if privateRegistry, ok := d.GetOk("private_registry"); ok && privateRegistry.(*schema.Set).Len() > 0 && privateRegistry.(*schema.Set).List()[0] != nil {
		secret.PrivateRegistry = expandPrivateRegistry(d.Get("private_registry").(*schema.Set).List())
	}

	if azureContainerRegistry, ok := d.GetOk("azure_container_registry"); ok && azureContainerRegistry.(*schema.Set).Len() > 0 && azureContainerRegistry.(*schema.Set).List()[0] != nil {
		secret.AzureContainerRegistry = expandAzureContainerRegistry(d.Get("azure_container_registry").(*schema.Set).List())
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

	if dockerHubRegistry, ok := d.GetOk("docker_hub_registry"); ok && dockerHubRegistry.(*schema.Set).Len() > 0 && dockerHubRegistry.(*schema.Set).List()[0] != nil {
		secret.DockerHubRegistry = expandDockerHubRegistry(dockerHubRegistry.(*schema.Set).List())
	}

	if gitHubRegistry, ok := d.GetOk("github_registry"); ok && gitHubRegistry.(*schema.Set).Len() > 0 && gitHubRegistry.(*schema.Set).List()[0] != nil {
		secret.GithubRegistry = expandGitHubRegistry(gitHubRegistry.(*schema.Set).List())
	}

	if doRegistry, ok := d.GetOk("digital_ocean_container_registry"); ok && doRegistry.(*schema.Set).Len() > 0 && doRegistry.(*schema.Set).List()[0] != nil {
		secret.DigitalOceanRegistry = expandDigitalOceanRegistry(doRegistry.(*schema.Set).List())
	}

	if gitLabRegistry, ok := d.GetOk("gitlab_registry"); ok && gitLabRegistry.(*schema.Set).Len() > 0 && gitLabRegistry.(*schema.Set).List()[0] != nil {
		secret.GitlabRegistry = expandGitLabRegistry(gitLabRegistry.(*schema.Set).List())
	}

	if privateRegistry, ok := d.GetOk("private_registry"); ok && privateRegistry.(*schema.Set).Len() > 0 && privateRegistry.(*schema.Set).List()[0] != nil {
		secret.PrivateRegistry = expandPrivateRegistry(privateRegistry.(*schema.Set).List())
	}

	if azureContainerRegistry, ok := d.GetOk("azure_container_registry"); ok && azureContainerRegistry.(*schema.Set).Len() > 0 && azureContainerRegistry.(*schema.Set).List()[0] != nil {
		secret.AzureContainerRegistry = expandAzureContainerRegistry(azureContainerRegistry.(*schema.Set).List())
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
