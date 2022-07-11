package koyeb

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

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

func expandGitHubRegistry(config []interface{}) *koyeb.GitHubRegistryConfiguration {
	rawGitHubRegistry := config[0].(map[string]interface{})

	gitHubRegistry := &koyeb.GitHubRegistryConfiguration{
		Username: toOpt(rawGitHubRegistry["username"].(string)),
		Password: toOpt(rawGitHubRegistry["password"].(string)),
	}

	return gitHubRegistry
}

func flattenGitHubRegistry(gitHubRegistry *koyeb.GitHubRegistryConfiguration) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = gitHubRegistry.GetUsername()
	r["password"] = gitHubRegistry.GetPassword()

	result = append(result, r)

	return result
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

func expandGitLabRegistry(config []interface{}) *koyeb.GitLabRegistryConfiguration {
	rawGitLabRegistry := config[0].(map[string]interface{})

	gitLabRegistry := &koyeb.GitLabRegistryConfiguration{
		Username: toOpt(rawGitLabRegistry["username"].(string)),
		Password: toOpt(rawGitLabRegistry["password"].(string)),
	}

	return gitLabRegistry
}

func flattenGitLabRegistry(gitLabRegistry *koyeb.GitLabRegistryConfiguration) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = gitLabRegistry.GetUsername()
	r["password"] = gitLabRegistry.GetPassword()

	result = append(result, r)

	return result
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

func expandDigitalOceanRegistry(config []interface{}) *koyeb.DigitalOceanRegistryConfiguration {
	rawDigitalOceanRegistry := config[0].(map[string]interface{})

	digitalOceanRegistry := &koyeb.DigitalOceanRegistryConfiguration{
		Username: toOpt(rawDigitalOceanRegistry["username"].(string)),
		Password: toOpt(rawDigitalOceanRegistry["password"].(string)),
	}

	return digitalOceanRegistry
}

func flattenDigitalOceanRegistry(digitalOceanRegistry *koyeb.DigitalOceanRegistryConfiguration) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = digitalOceanRegistry.GetUsername()
	r["password"] = digitalOceanRegistry.GetPassword()

	result = append(result, r)

	return result
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

func expandDockerHubRegistry(config []interface{}) *koyeb.DockerHubRegistryConfiguration {
	rawDockerHubRegistry := config[0].(map[string]interface{})

	dockerHubRegistry := &koyeb.DockerHubRegistryConfiguration{
		Username: toOpt(rawDockerHubRegistry["username"].(string)),
		Password: toOpt(rawDockerHubRegistry["password"].(string)),
	}

	return dockerHubRegistry
}

func flattenDockerHubRegistry(dockerHubRegistry *koyeb.DockerHubRegistryConfiguration) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = dockerHubRegistry.GetUsername()
	r["password"] = dockerHubRegistry.GetPassword()

	result = append(result, r)

	return result
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
				// ValidateFunc: validation.IsURLWithScheme([]string{"https", "http"}),
			},
		},
	}
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

func flattenPrivateRegistry(privateRegistry *koyeb.PrivateRegistryConfiguration) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = privateRegistry.GetUsername()
	r["password"] = privateRegistry.GetPassword()
	r["url"] = privateRegistry.GetUrl()

	result = append(result, r)

	return result
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

func expandAzureContainerRegistry(config []interface{}) *koyeb.AzureContainerRegistryConfiguration {
	rawAzureContainerRegistry := config[0].(map[string]interface{})

	azureContainerRegistry := &koyeb.AzureContainerRegistryConfiguration{
		Username:     toOpt(rawAzureContainerRegistry["username"].(string)),
		Password:     toOpt(rawAzureContainerRegistry["password"].(string)),
		RegistryName: toOpt(rawAzureContainerRegistry["registry_name"].(string)),
	}

	return azureContainerRegistry
}

func flattenAzureContainerRegistry(azureContainerRegistry *koyeb.AzureContainerRegistryConfiguration) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["username"] = azureContainerRegistry.GetUsername()
	r["password"] = azureContainerRegistry.GetPassword()
	r["registry_name"] = azureContainerRegistry.GetRegistryName()

	result = append(result, r)

	return result
}

func secretSchema() map[string]*schema.Schema {
	secret := map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Description:  "The secret name",
			ForceNew:     true,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(2, 64),
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
				"azure_container_registry",
				"github_registry",
				"gitlab_registry",
				"private_registry",
				"digital_ocean_container_registry",
			},
		},
		"docker_hub_registry": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        dockerHubRegistrySchema(),
			Description: "The DockerHub registry configuration to use",
			MaxItems:    1,
			ConflictsWith: []string{
				"azure_container_registry",
				"github_registry",
				"gitlab_registry",
				"private_registry",
				"digital_ocean_container_registry",
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
				"azure_container_registry",
				"gitlab_registry",
				"private_registry",
				"digital_ocean_container_registry",
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
				"azure_container_registry",
				"github_registry",
				"private_registry",
				"digital_ocean_container_registry",
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
				"azure_container_registry",
				"github_registry",
				"gitlab_registry",
				"private_registry",
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
				"azure_container_registry",
				"github_registry",
				"gitlab_registry",
				"digital_ocean_container_registry",
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
				"private_registry",
				"digital_ocean_container_registry",
			},
			Set: schema.HashResource(privateRegistrySchema()),
		},
		"organization_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The organization id owning the app",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the Secret was last updated",
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The date and time of when the Secret was created",
		},
	}

	return secret
}

func resourceKoyebSecret() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Secret resource in the Koyeb Terraform provider.",

		CreateContext: resourceKoyebSecretCreate,
		ReadContext:   resourceKoyebSecretRead,
		UpdateContext: resourceKoyebSecretUpdate,
		DeleteContext: resourceKoyebSecretDelete,

		Schema: secretSchema(),
	}
}

func setSecretAttribute(d *schema.ResourceData, secret koyeb.Secret) error {
	d.SetId(secret.GetId())
	d.Set("name", secret.GetName())
	d.Set("type", secret.GetType())
	// d.Set("value", secret.GetValue())
	// d.Set("docker_hub_registry", flattenDockerHubRegistry(secret.DockerHubRegistry))
	// d.Set("github_registry", flattenGitHubRegistry(secret.GithubRegistry))
	// d.Set("gitlab_registry", flattenGitLabRegistry(secret.GitlabRegistry))
	// d.Set("digital_ocean_container_registry", flattenDigitalOceanRegistry(secret.DigitalOceanRegistry))
	// d.Set("private_registry", flattenPrivateRegistry(secret.PrivateRegistry))
	// d.Set("azure_container_registry", flattenAzureContainerRegistry(secret.AzureContainerRegistry))
	d.Set("organization_id", secret.GetOrganizationId())
	d.Set("created_at", secret.GetCreatedAt().UTC().String())
	d.Set("updated_at", secret.GetUpdatedAt().UTC().String())

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

	res, resp, err := client.SecretsApi.CreateSecret(ctx).Body(secret).Execute()
	if err != nil {
		return diag.Errorf("Error creating secret: %s (%v %v)", err, resp, res)
	}

	d.SetId(*res.Secret.Id)
	log.Printf("[INFO] Created secret name: %s", *res.Secret.Name)

	return resourceKoyebSecretRead(ctx, d, meta)
}

func resourceKoyebSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*koyeb.APIClient)

	res, resp, err := client.SecretsApi.GetSecret(context.Background(), d.Id()).Execute()
	if err != nil {
		// If the Secret is somehow already destroyed, mark as
		// successfully gone
		if resp.StatusCode == 404 {
			d.SetId("")
			return nil
		}

		return diag.Errorf("Error retrieving secret: %s (%v %v)", err, resp, res)
	}

	setSecretAttribute(d, *res.Secret)

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

	res, resp, err := client.SecretsApi.UpdateSecret(context.Background(), d.Id()).Body(secret).Execute()

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
