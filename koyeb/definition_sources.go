package koyeb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func archiveSourceSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The archive ID to deploy, as uploaded by `koyeb deploy`",
			},
			"buildpack": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     buildpackBuilderSchema(),
				Set:      schema.HashResource(buildpackBuilderSchema()),
				MaxItems: 1,
			},
			"dockerfile": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     dockerBuilderSchema(),
				Set:      schema.HashResource(dockerBuilderSchema()),
				MaxItems: 1,
			},
		},
	}
}

func expandArchiveSource(config []interface{}) *koyeb.ArchiveSource {
	rawArchiveSource := config[0].(map[string]interface{})

	archiveSource := &koyeb.ArchiveSource{
		Id: toOpt(rawArchiveSource["id"].(string)),
	}

	if rawArchiveSource["dockerfile"] != nil && rawArchiveSource["dockerfile"].(*schema.Set).Len() > 0 {
		archiveSource.Docker = expandDockerBuilder(rawArchiveSource["dockerfile"].(*schema.Set).List())
	} else if rawArchiveSource["buildpack"] != nil && rawArchiveSource["buildpack"].(*schema.Set).Len() > 0 {
		archiveSource.Buildpack = expandBuildpackBuilder(rawArchiveSource["buildpack"].(*schema.Set).List())
	}

	return archiveSource
}

func flattenArchive(archiveSource *koyeb.ArchiveSource) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["id"] = archiveSource.GetId()
	if buildpack, ok := archiveSource.GetBuildpackOk(); ok {
		r["buildpack"] = flattenBuildpackBuilder(buildpack)
	}
	if docker, ok := archiveSource.GetDockerOk(); ok {
		r["dockerfile"] = flattenDockerBuilder(docker)
	}

	result = append(result, r)

	return result
}

func dockerSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"image": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Docker image to use to support your service",
			},
			"command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Docker command to use",
			},
			"args": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Docker args to use",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"entrypoint": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The Docker entrypoint to use",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"privileged": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When enabled, the service container will run in privileged mode. This advanced feature is useful to get advanced system privileges.",
			},
			"image_registry_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Koyeb secret containing the container registry credentials",
			},
		},
	}
}

func gitSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"repository": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The GitHub repository to deploy",
			},
			"branch": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The GitHub branch to deploy. Exactly one of branch, tag or sha must be set.",
			},
			"tag": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The GitHub repository tag to deploy. Exactly one of branch, tag or sha must be set.",
			},
			"sha": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The git commit SHA to deploy. Exactly one of branch, tag or sha must be set.",
			},
			"workdir": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The directory where your source code is located. If not set, the work directory defaults to the root of the repository.",
			},
			"buildpack": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     buildpackBuilderSchema(),
				Set:      schema.HashResource(buildpackBuilderSchema()),
				MaxItems: 1,
			},
			"dockerfile": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     dockerBuilderSchema(),
				Set:      schema.HashResource(dockerBuilderSchema()),
				MaxItems: 1,
			},
			"no_deploy_on_push": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set to true, no Koyeb deployments will be triggered when changes are pushed to the GitHub repository branch",
			},
		},
	}
}

func buildpackBuilderSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"build_command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The command to build your application during the build phase. If your application does not require a build command, leave this field empty",
			},
			"run_command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The command to run your application once the built is completed",
			},
			"privileged": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When enabled, the service container will run in privileged mode. This advanced feature is useful to get advanced system privileges.",
			},
		},
	}
}

func dockerBuilderSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dockerfile": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The location of your Dockerfile relative to the work directory. If not set, the work directory defaults to the root of the repository.",
			},
			"entrypoint": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Override the default entrypoint to execute on the container",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"command": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Override the command to execute on the container",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"args": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The arguments to pass to the Docker command",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"target": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Target build stage: If your Dockerfile contains multi-stage builds, you can choose the target stage to build and deploy by entering its name",
			},
			"privileged": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "When enabled, the service container will run in privileged mode. This advanced feature is useful to get advanced system privileges.",
			},
		},
	}
}

func databaseSourceSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"neon_postgres": {
				Type:        schema.TypeSet,
				Optional:    true,
				MaxItems:    1,
				Description: "The Neon PostgreSQL database to provision",
				Elem:        neonPostgresSchema(),
				Set:         schema.HashResource(neonPostgresSchema()),
			},
		},
	}
}

func neonPostgresSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"pg_version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The PostgreSQL version",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region where the database is deployed",
			},
			"instance_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The instance type of the database (free, small, medium or large)",
			},
			"databases": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The databases to create in the PostgreSQL instance",
				Elem:        neonDatabaseSchema(),
				Set:         schema.HashResource(neonDatabaseSchema()),
			},
			"roles": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The roles to create in the PostgreSQL instance",
				Elem:        neonRoleSchema(),
				Set:         schema.HashResource(neonRoleSchema()),
			},
		},
	}
}

func neonDatabaseSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The database name",
			},
			"owner": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The role owning the database",
			},
		},
	}
}

func neonRoleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The role name",
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the managed secret holding the role password",
			},
		},
	}
}

// configFileSchema is shared between the service resource and data source; the
// pattern is package-level so repeated schema construction does not
// recompile it.

func flattenDatabase(database *koyeb.DatabaseSource) []interface{} {
	result := make([]interface{}, 0)

	if database.NeonPostgres == nil {
		return result
	}

	neon := database.NeonPostgres

	neonMap := map[string]interface{}{
		"pg_version":    int(neon.GetPgVersion()),
		"region":        neon.GetRegion(),
		"instance_type": neon.GetInstanceType(),
		"databases": schema.NewSet(
			schema.HashResource(neonDatabaseSchema()),
			flattenNeonDatabases(neon.Databases),
		),
		"roles": schema.NewSet(
			schema.HashResource(neonRoleSchema()),
			flattenNeonRoles(neon.Roles),
		),
	}

	result = append(result, map[string]interface{}{
		"neon_postgres": schema.NewSet(
			schema.HashResource(neonPostgresSchema()),
			[]interface{}{neonMap},
		),
	})

	return result
}

func flattenNeonDatabases(databases []koyeb.NeonPostgresDatabaseNeonDatabase) []interface{} {
	result := make([]interface{}, len(databases))
	for i, database := range databases {
		result[i] = map[string]interface{}{
			"name":  database.GetName(),
			"owner": database.GetOwner(),
		}
	}
	return result
}

func flattenNeonRoles(roles []koyeb.NeonPostgresDatabaseNeonRole) []interface{} {
	result := make([]interface{}, len(roles))
	for i, role := range roles {
		result[i] = map[string]interface{}{
			"name":   role.GetName(),
			"secret": role.GetSecret(),
		}
	}
	return result
}

func expandDatabaseSource(config []interface{}) *koyeb.DatabaseSource {
	if len(config) == 0 {
		return nil
	}

	rawDatabase := config[0].(map[string]interface{})
	databaseSource := &koyeb.DatabaseSource{}

	neonPostgres := rawDatabase["neon_postgres"].(*schema.Set).List()
	if len(neonPostgres) == 0 {
		return databaseSource
	}

	rawNeonPostgres := neonPostgres[0].(map[string]interface{})
	neon := koyeb.NeonPostgresDatabase{
		PgVersion:    toOpt(int64(rawNeonPostgres["pg_version"].(int))),
		Region:       toOpt(rawNeonPostgres["region"].(string)),
		InstanceType: toOpt(rawNeonPostgres["instance_type"].(string)),
	}

	for _, rawDatabaseItem := range rawNeonPostgres["databases"].(*schema.Set).List() {
		databaseItem := rawDatabaseItem.(map[string]interface{})
		neon.Databases = append(neon.Databases, koyeb.NeonPostgresDatabaseNeonDatabase{
			Name:  toOpt(databaseItem["name"].(string)),
			Owner: toOpt(databaseItem["owner"].(string)),
		})
	}

	for _, rawRole := range rawNeonPostgres["roles"].(*schema.Set).List() {
		role := rawRole.(map[string]interface{})
		neon.Roles = append(neon.Roles, koyeb.NeonPostgresDatabaseNeonRole{
			Name:   toOpt(role["name"].(string)),
			Secret: toOpt(role["secret"].(string)),
		})
	}

	databaseSource.NeonPostgres = &neon

	return databaseSource
}

func expandDockerSource(config []interface{}) *koyeb.DockerSource {
	rawDockerSource := config[0].(map[string]interface{})

	dockerSource := &koyeb.DockerSource{
		Image: toOpt(rawDockerSource["image"].(string)),
	}

	if rawDockerSource["command"] != nil {
		dockerSource.Command = toOpt(rawDockerSource["command"].(string))
	}

	rawArgs := rawDockerSource["args"].([]interface{})
	args := make([]string, len(rawArgs))
	for i, v := range rawArgs {
		args[i] = v.(string)
	}
	dockerSource.Args = args

	rawEntrypoint := rawDockerSource["entrypoint"].([]interface{})
	entrypoint := make([]string, len(rawEntrypoint))
	for i, v := range rawEntrypoint {
		entrypoint[i] = v.(string)
	}
	dockerSource.Entrypoint = entrypoint

	if rawDockerSource["privileged"] != nil {
		dockerSource.Privileged = toOpt(rawDockerSource["privileged"].(bool))
	}

	if rawDockerSource["image_registry_secret"] != nil {
		dockerSource.ImageRegistrySecret = toOpt(rawDockerSource["image_registry_secret"].(string))
	}

	return dockerSource
}

func flattenDocker(dockerSource *koyeb.DockerSource) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["image"] = dockerSource.Image
	r["command"] = dockerSource.Command
	r["args"] = dockerSource.Args
	r["entrypoint"] = dockerSource.Entrypoint
	r["privileged"] = dockerSource.Privileged
	r["image_registry_secret"] = dockerSource.ImageRegistrySecret

	result = append(result, r)

	return result
}

func expandDockerBuilder(config []interface{}) *koyeb.DockerBuilder {
	rawDockerBuilderSource := config[0].(map[string]interface{})

	dockerBuilderSource := &koyeb.DockerBuilder{}

	if rawDockerBuilderSource["dockerfile"] != nil {
		dockerBuilderSource.Dockerfile = toOpt(rawDockerBuilderSource["dockerfile"].(string))
	}

	rawEntrypoint := rawDockerBuilderSource["entrypoint"].([]interface{})
	entrypoint := make([]string, len(rawEntrypoint))
	for i, v := range rawEntrypoint {
		entrypoint[i] = v.(string)
	}
	dockerBuilderSource.Entrypoint = entrypoint

	if rawDockerBuilderSource["command"] != nil {
		dockerBuilderSource.Command = toOpt(rawDockerBuilderSource["command"].(string))
	}

	rawArgs := rawDockerBuilderSource["args"].([]interface{})
	args := make([]string, len(rawArgs))
	for i, v := range rawArgs {
		args[i] = v.(string)
	}
	dockerBuilderSource.Args = args

	if rawDockerBuilderSource["target"] != nil {
		dockerBuilderSource.Target = toOpt(rawDockerBuilderSource["target"].(string))
	}

	if rawDockerBuilderSource["privileged"] != nil {
		dockerBuilderSource.Privileged = toOpt(rawDockerBuilderSource["privileged"].(bool))
	}

	return dockerBuilderSource
}

func flattenDockerBuilder(dockerBuilderSource *koyeb.DockerBuilder) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["entrypoint"] = dockerBuilderSource.Entrypoint
	r["command"] = dockerBuilderSource.Command
	r["args"] = dockerBuilderSource.Args
	r["target"] = dockerBuilderSource.Target
	r["privileged"] = dockerBuilderSource.Privileged

	result = append(result, r)

	return result
}

func expandBuildpackBuilder(config []interface{}) *koyeb.BuildpackBuilder {
	rawBuildpackBuilderSource := config[0].(map[string]interface{})

	buildpackBuilderSource := &koyeb.BuildpackBuilder{}

	if rawBuildpackBuilderSource["build_command"] != nil {
		buildpackBuilderSource.BuildCommand = toOpt(rawBuildpackBuilderSource["build_command"].(string))
	}

	if rawBuildpackBuilderSource["run_command"] != nil {
		buildpackBuilderSource.RunCommand = toOpt(rawBuildpackBuilderSource["run_command"].(string))
	}

	if rawBuildpackBuilderSource["privileged"] != nil {
		buildpackBuilderSource.Privileged = toOpt(rawBuildpackBuilderSource["privileged"].(bool))
	}

	return buildpackBuilderSource
}

func flattenBuildpackBuilder(buildpackBuilderSource *koyeb.BuildpackBuilder) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["build_command"] = buildpackBuilderSource.GetBuildCommand()
	r["run_command"] = buildpackBuilderSource.GetRunCommand()
	r["privileged"] = buildpackBuilderSource.GetPrivileged()

	result = append(result, r)

	return result
}

func expandGitSource(config []interface{}) *koyeb.GitSource {
	rawGitSource := config[0].(map[string]interface{})

	gitSource := &koyeb.GitSource{
		Repository:     toOpt(rawGitSource["repository"].(string)),
		Branch:         toOpt(rawGitSource["branch"].(string)),
		Tag:            toOpt(rawGitSource["tag"].(string)),
		Sha:            toOpt(rawGitSource["sha"].(string)),
		Workdir:        toOpt(rawGitSource["workdir"].(string)),
		NoDeployOnPush: toOpt(rawGitSource["no_deploy_on_push"].(bool)),
	}

	if rawGitSource["dockerfile"] != nil && rawGitSource["dockerfile"].(*schema.Set).Len() > 0 {
		gitSource.Docker = expandDockerBuilder(rawGitSource["dockerfile"].(*schema.Set).List())
	} else if rawGitSource["buildpack"] != nil && rawGitSource["buildpack"].(*schema.Set).Len() > 0 {
		gitSource.Buildpack = expandBuildpackBuilder(rawGitSource["buildpack"].(*schema.Set).List())
	}

	return gitSource
}

func flattenGit(gitSource *koyeb.GitSource) []interface{} {
	result := make([]interface{}, 0)

	r := make(map[string]interface{})
	r["repository"] = gitSource.GetRepository()
	r["branch"] = gitSource.GetBranch()
	r["tag"] = gitSource.GetTag()
	r["sha"] = gitSource.GetSha()
	r["workdir"] = gitSource.GetWorkdir()
	r["no_deploy_on_push"] = gitSource.GetNoDeployOnPush()
	if buildpack, ok := gitSource.GetBuildpackOk(); ok {
		r["buildpack"] = flattenBuildpackBuilder(buildpack)
	}
	if docker, ok := gitSource.GetDockerOk(); ok {
		r["dockerfile"] = flattenDockerBuilder(docker)
	}

	result = append(result, r)

	return result
}
