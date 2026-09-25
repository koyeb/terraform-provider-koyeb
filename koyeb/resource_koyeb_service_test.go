package koyeb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func init() {
	resource.AddTestSweepers("koyeb_service", &resource.Sweeper{
		Name:         "koyeb_service",
		F:            testSweepService,
		Dependencies: []string{"koyeb_app"},
	})

}

func testSweepService(string) error {
	meta, err := sharedConfig()
	if err != nil {
		return err
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.ServicesApi.ListServices(context.Background()).Limit("100").Execute()
	if err != nil {
		return err
	}

	for _, a := range res.Services {
		if strings.HasPrefix(a.GetName(), testNamePrefix) {
			log.Printf("Destroying service %s", *a.Name)

			if _, _, err := client.ServicesApi.DeleteService(context.Background(), a.GetId()).Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccKoyebService_Basic(t *testing.T) {
	var service koyeb.Service
	appName := randomTestName()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_basic_docker, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "service"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "created_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "app_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "version"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "messages"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "paused_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "resumed_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "terminated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "latest_deployment"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_basic_git_buildpack, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "service"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "created_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "app_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "version"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "messages"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "paused_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "resumed_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "terminated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "latest_deployment"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_basic_git_dockerfile, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "service"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "created_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "app_id"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "version"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "status"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "messages"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "paused_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "resumed_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "terminated_at"),
					resource.TestCheckResourceAttrSet("koyeb_service.bar", "latest_deployment"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_docker_sleep_idle_delay, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "service"),
					resource.TestCheckTypeSetElemNestedAttrs("koyeb_service.bar", "definition.0.scalings.*.targets.*.sleep_idle_delay.*", map[string]string{
						"light_sleep_value": "60",
						"deep_sleep_value":  "300",
					}),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_docker_strategy_proxyports_configfiles, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "service"),
					resource.TestCheckResourceAttr("koyeb_service.bar", "definition.0.strategy", "DEPLOYMENT_STRATEGY_TYPE_BLUE_GREEN"),
					resource.TestCheckTypeSetElemNestedAttrs("koyeb_service.bar", "definition.0.proxy_ports.*", map[string]string{
						"port":     "22",
						"protocol": "tcp",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("koyeb_service.bar", "definition.0.config_files.*", map[string]string{
						"path":        "/etc/data.yaml",
						"content":     "key: value",
						"permissions": "0644",
					}),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebServiceConfig_docker_mesh_network_auth, appName, appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebServiceExists("koyeb_service.bar", &service),
					resource.TestCheckResourceAttr("koyeb_service.bar", "name", "service"),
					resource.TestCheckResourceAttr("koyeb_service.bar", "definition.0.mesh", "DEPLOYMENT_MESH_DISABLED"),
					resource.TestCheckTypeSetElemNestedAttrs("koyeb_service.bar", "definition.0.network_policy.*.egress.*", map[string]string{
						"mode": "EGRESS_POLICY_MODE_DEFAULT",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("koyeb_service.bar", "definition.0.routes.*.security_policies.*.basic_auths.*", map[string]string{
						"username": "user",
						"password": "password",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("koyeb_service.bar", "definition.0.routes.*.security_policies.*", map[string]string{
						"api_keys.#": "1",
					}),
				),
			},
		},
	})
}

func testAccCheckKoyebServiceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"DELETED", "DELETING"}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_service" {
			continue
		}

		err := waitForStatus(context.Background(), goneWait("Service", targetStatus, time.Minute), serviceStatusPoller(context.Background(), client, rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("Service still exists: %s ", err)
		}

	}

	return nil
}

func testAccCheckKoyebServiceExists(n string, service *koyeb.Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.ServicesApi.GetService(context.Background(), rs.Primary.ID).Execute()

		if err != nil {
			return err
		}

		if *res.Service.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		a := res.GetService()
		*service = a

		return nil
	}
}

const testAccCheckKoyebServiceConfig_docker_mesh_network_auth = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "service"
		instance_types {
		  type = "micro"
		}
		ports {
		  port     = 3000
		  protocol = "http"
		}
		scalings {
		  min = 1
		  max = 1
		}
		mesh = "DEPLOYMENT_MESH_DISABLED"
		network_policy {
		  egress {
		    mode = "EGRESS_POLICY_MODE_DEFAULT"
		  }
		}
		routes {
		  port = 3000
		  path = "/"
		  security_policies {
		    basic_auths {
		      username = "user"
		      password = "password"
		    }
		    api_keys = ["api-key"]
		  }
		}
		regions = ["tyo"]
		docker {
		  image = "koyeb/demo"
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`

const testAccCheckKoyebServiceConfig_docker_strategy_proxyports_configfiles = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "service"
		instance_types {
		  type = "micro"
		}
		ports {
		  port     = 3000
		  protocol = "http"
		}
		scalings {
		  min = 1
		  max = 1
		}
		strategy = "DEPLOYMENT_STRATEGY_TYPE_BLUE_GREEN"
		proxy_ports {
		  port     = 22
		  protocol = "tcp"
		}
		config_files {
		  path        = "/etc/data.yaml"
		  content     = "key: value"
		  permissions = "0644"
		}
		regions = ["tyo"]
		docker {
		  image = "koyeb/demo"
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`

const testAccCheckKoyebServiceConfig_basic_docker = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "service"
		instance_types {
		  type = "micro"
		}
		ports {
		  port     = 3000
		  protocol = "http"
		}
		scalings {
		  min = 1
		  max = 1
		}
		env {
		  key   = "FOO"
		  value = "BAR"
		}
		routes {
		  path = "/"
		  port = 3000
		}
		health_checks {
		  http {
		    path = "/"
		    port = 3000
			headers {
				key = "X-Header"
				value = "Value"
			}
		  }
		}
		regions = ["tyo"]
		docker {
		  image = "koyeb/demo"
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`

const testAccCheckKoyebServiceConfig_docker_sleep_idle_delay = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "service"
		instance_types {
		  type = "micro"
		}
		ports {
		  port     = 3000
		  protocol = "http"
		}
		scalings {
		  min = 0
		  max = 1
		  targets {
		    sleep_idle_delay {
		      light_sleep_value = 60
		      deep_sleep_value  = 300
		    }
		  }
		}
		routes {
		  port = 3000
		  path = "/"
		}
		regions = ["tyo"]
		docker {
		  image = "koyeb/demo"
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`

const testAccCheckKoyebServiceConfig_basic_git_buildpack = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "service"
		instance_types {

		  type = "micro"
		}
		ports {
		  port     = 8080
		  protocol = "http"
		}
		scalings {
		  min = 1
		  max = 1
		}
		env {
		  key   = "FOO"
		  value = "BAR"
		}
		// The Procfile runs gunicorn on $PORT: without this env the app
		// cannot bind the health-checked port and the deployment errors.
		env {
		  key   = "PORT"
		  value = "8080"
		}
		routes {
		  path = "/"
		  port = 8080
		}
		health_checks {
		  tcp {
			port = 8080
		  }
		}
		regions = ["fra"]
		git {
		  repository = "github.com/koyeb/example-flask"
		  branch = "main"
		  buildpack {}
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`

const testAccCheckKoyebServiceConfig_basic_git_dockerfile = `
resource "koyeb_app" "foo" {
	name = "%s"
}

resource "koyeb_service" "bar" {
	app_name = "%s"
	definition {
		name = "service"
		instance_types {

		  type = "micro"
		}
		// The platform requires a route on services scaling to zero
		// (workers cannot), so expose the express app's default port.
		ports {
		  port     = 3000
		  protocol = "http"
		}
		routes {
		  path = "/"
		  port = 3000
		}
		scalings {
		  min = 0
		  max = 1
		}
		env {
		  key   = "FOO"
		  value = "BAR"
		}
		regions = ["fra", "tyo"]
		git {
		  // example-flask has no Dockerfile and the platform e2e's
		  // docker-build fixture is private (the org's GitHub
		  // integration cannot resolve its SHA), so build the public
		  // express example instead.
		  repository = "github.com/koyeb/example-expressjs"
		  branch = "main"
		  dockerfile {}
		}
	}

	depends_on = [
	  koyeb_app.foo
	]
}`

func TestExpandGitSourceSetsTagAndSha(t *testing.T) {
	config := []interface{}{
		map[string]interface{}{
			"repository":        "github.com/koyeb/example",
			"branch":            "",
			"tag":               "v1.2.3",
			"sha":               "0123456789abcdef",
			"workdir":           "",
			"no_deploy_on_push": false,
		},
	}

	gitSource := expandGitSource(config)

	if gitSource.GetTag() != "v1.2.3" {
		t.Errorf("expected tag %q, got %q", "v1.2.3", gitSource.GetTag())
	}
	if gitSource.GetSha() != "0123456789abcdef" {
		t.Errorf("expected sha %q, got %q", "0123456789abcdef", gitSource.GetSha())
	}
}

func TestFlattenGitSetsTagAndSha(t *testing.T) {
	gitSource := &koyeb.GitSource{
		Repository: toOpt("github.com/koyeb/example"),
		Branch:     toOpt("main"),
		Tag:        toOpt("v1.2.3"),
		Sha:        toOpt("0123456789abcdef"),
	}

	flattened := flattenGit(gitSource)[0].(map[string]interface{})

	if flattened["tag"] != "v1.2.3" {
		t.Errorf("expected tag %q, got %v", "v1.2.3", flattened["tag"])
	}
	if flattened["sha"] != "0123456789abcdef" {
		t.Errorf("expected sha %q, got %v", "0123456789abcdef", flattened["sha"])
	}
}

func TestExpandScalingsSetsSleepIdleDelay(t *testing.T) {
	sleepIdleDelay := schema.NewSet(func(_ interface{}) int { return 0 }, []interface{}{
		map[string]interface{}{"light_sleep_value": 300, "deep_sleep_value": 1800},
	})
	targets := schema.NewSet(func(_ interface{}) int { return 0 }, []interface{}{
		map[string]interface{}{"sleep_idle_delay": sleepIdleDelay},
	})
	config := []interface{}{
		map[string]interface{}{
			"min":     0,
			"max":     1,
			"scopes":  []interface{}{},
			"targets": targets,
		},
	}

	scalings := expandScalings(config)

	if len(scalings) != 1 {
		t.Fatalf("expected 1 scaling, got %d", len(scalings))
	}

	var sleepIdleDelayTarget *koyeb.DeploymentScalingTargetSleepIdleDelay
	for _, target := range scalings[0].Targets {
		if target.SleepIdleDelay != nil {
			sleepIdleDelayTarget = target.SleepIdleDelay
		}
	}
	if sleepIdleDelayTarget == nil {
		t.Fatal("expected a sleep_idle_delay scaling target")
	}
	if sleepIdleDelayTarget.GetLightSleepValue() != 300 {
		t.Errorf("expected light_sleep_value 300, got %d", sleepIdleDelayTarget.GetLightSleepValue())
	}
	if sleepIdleDelayTarget.GetDeepSleepValue() != 1800 {
		t.Errorf("expected deep_sleep_value 1800, got %d", sleepIdleDelayTarget.GetDeepSleepValue())
	}
}

func TestFlattenScalingsSetsSleepIdleDelay(t *testing.T) {
	scalings := []koyeb.DeploymentScaling{
		{
			Min: toOpt(int64(0)),
			Max: toOpt(int64(1)),
			Targets: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: toOpt(int64(300)),
						DeepSleepValue:  toOpt(int64(1800)),
					},
				},
			},
		},
	}

	flattened := flattenScalings(&scalings)

	targets := flattened[0]["targets"].(*schema.Set).List()
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	target := targets[0].(map[string]interface{})

	sleepIdleDelaySet, ok := target["sleep_idle_delay"].(*schema.Set)
	if !ok {
		t.Fatal("expected a sleep_idle_delay set in the flattened target")
	}
	sleepIdleDelay := sleepIdleDelaySet.List()[0].(map[string]interface{})
	if sleepIdleDelay["light_sleep_value"] != 300 {
		t.Errorf("expected light_sleep_value 300, got %v", sleepIdleDelay["light_sleep_value"])
	}
	if sleepIdleDelay["deep_sleep_value"] != 1800 {
		t.Errorf("expected deep_sleep_value 1800, got %v", sleepIdleDelay["deep_sleep_value"])
	}
}

func testRawDefinition() map[string]interface{} {
	emptySet := func() *schema.Set { return schema.NewSet(func(_ interface{}) int { return 0 }, nil) }

	return map[string]interface{}{
		"name":           "service",
		"type":           "WEB",
		"env":            emptySet(),
		"ports":          emptySet(),
		"routes":         emptySet(),
		"scalings":       emptySet(),
		"instance_types": emptySet(),
		"regions":        emptySet(),
		"health_checks":  emptySet(),
		"volumes":        emptySet(),
		"proxy_ports":    emptySet(),
		"config_files":   emptySet(),
		"database":       emptySet(),
		"network_policy": emptySet(),
		"archive":        emptySet(),
		"git":            emptySet(),
		"docker":         emptySet(),
		"skip_cache":     false,
	}
}

func TestExpandDeploymentDefinitionSetsStrategy(t *testing.T) {
	raw := testRawDefinition()
	raw["strategy"] = "DEPLOYMENT_STRATEGY_TYPE_BLUE_GREEN"

	definition := expandDeploymentDefinition(raw)

	if definition.Strategy == nil {
		t.Fatal("expected a strategy to be set")
	}
	if definition.Strategy.GetType() != koyeb.DEPLOYMENTSTRATEGYTYPE_BLUE_GREEN {
		t.Errorf("expected strategy %q, got %q", koyeb.DEPLOYMENTSTRATEGYTYPE_BLUE_GREEN, definition.Strategy.GetType())
	}
}

func TestFlattenDeploymentDefinitionSetsStrategy(t *testing.T) {
	definition := &koyeb.DeploymentDefinition{
		Strategy: &koyeb.DeploymentStrategy{
			Type: toOpt(koyeb.DEPLOYMENTSTRATEGYTYPE_BLUE_GREEN),
		},
	}

	flattened := flattenDeploymentDefinition(definition)[0].(map[string]interface{})

	if flattened["strategy"] != string(koyeb.DEPLOYMENTSTRATEGYTYPE_BLUE_GREEN) {
		t.Errorf("expected strategy %q, got %v", koyeb.DEPLOYMENTSTRATEGYTYPE_BLUE_GREEN, flattened["strategy"])
	}
}

func TestExpandDeploymentDefinitionSetsProxyPorts(t *testing.T) {
	raw := testRawDefinition()
	raw["proxy_ports"] = schema.NewSet(func(_ interface{}) int { return 0 }, []interface{}{
		map[string]interface{}{"port": 22, "protocol": "tcp"},
	})

	definition := expandDeploymentDefinition(raw)

	if len(definition.ProxyPorts) != 1 {
		t.Fatalf("expected 1 proxy port, got %d", len(definition.ProxyPorts))
	}
	if definition.ProxyPorts[0].GetPort() != 22 {
		t.Errorf("expected port 22, got %d", definition.ProxyPorts[0].GetPort())
	}
	if definition.ProxyPorts[0].GetProtocol() != koyeb.PROXYPORTPROTOCOL_TCP {
		t.Errorf("expected protocol %q, got %q", koyeb.PROXYPORTPROTOCOL_TCP, definition.ProxyPorts[0].GetProtocol())
	}
}

func TestFlattenDeploymentDefinitionSetsProxyPorts(t *testing.T) {
	definition := &koyeb.DeploymentDefinition{
		ProxyPorts: []koyeb.DeploymentProxyPort{
			{
				Port:     toOpt(int64(22)),
				Protocol: toOpt(koyeb.PROXYPORTPROTOCOL_TCP),
			},
		},
	}

	flattened := flattenDeploymentDefinition(definition)[0].(map[string]interface{})

	proxyPorts, ok := flattened["proxy_ports"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected proxy_ports to be flattened, got %T", flattened["proxy_ports"])
	}
	if len(proxyPorts) != 1 {
		t.Fatalf("expected 1 proxy port, got %d", len(proxyPorts))
	}
	if proxyPorts[0]["port"] != 22 {
		t.Errorf("expected port 22, got %v", proxyPorts[0]["port"])
	}
	if proxyPorts[0]["protocol"] != "tcp" {
		t.Errorf("expected protocol tcp, got %v", proxyPorts[0]["protocol"])
	}
}

func TestExpandDeploymentDefinitionSetsConfigFiles(t *testing.T) {
	raw := testRawDefinition()
	raw["config_files"] = schema.NewSet(func(_ interface{}) int { return 0 }, []interface{}{
		map[string]interface{}{
			"path":        "/etc/data.yaml",
			"content":     "key: value",
			"permissions": "0644",
		},
	})

	definition := expandDeploymentDefinition(raw)

	if len(definition.ConfigFiles) != 1 {
		t.Fatalf("expected 1 config file, got %d", len(definition.ConfigFiles))
	}
	if definition.ConfigFiles[0].GetPath() != "/etc/data.yaml" {
		t.Errorf("expected path %q, got %q", "/etc/data.yaml", definition.ConfigFiles[0].GetPath())
	}
	if definition.ConfigFiles[0].GetContent() != "key: value" {
		t.Errorf("expected content to be set, got %q", definition.ConfigFiles[0].GetContent())
	}
	if definition.ConfigFiles[0].GetPermissions() != "0644" {
		t.Errorf("expected permissions 0644, got %q", definition.ConfigFiles[0].GetPermissions())
	}
}

func TestFlattenDeploymentDefinitionSetsConfigFiles(t *testing.T) {
	definition := &koyeb.DeploymentDefinition{
		ConfigFiles: []koyeb.ConfigFile{
			{
				Path:        toOpt("/etc/data.yaml"),
				Content:     toOpt("key: value"),
				Permissions: toOpt("0644"),
			},
		},
	}

	flattened := flattenDeploymentDefinition(definition)[0].(map[string]interface{})

	configFiles, ok := flattened["config_files"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected config_files to be flattened, got %T", flattened["config_files"])
	}
	if len(configFiles) != 1 {
		t.Fatalf("expected 1 config file, got %d", len(configFiles))
	}
	if configFiles[0]["path"] != "/etc/data.yaml" {
		t.Errorf("expected path %q, got %v", "/etc/data.yaml", configFiles[0]["path"])
	}
	if configFiles[0]["content"] != "key: value" {
		t.Errorf("expected content to be set, got %v", configFiles[0]["content"])
	}
	if configFiles[0]["permissions"] != "0644" {
		t.Errorf("expected permissions 0644, got %v", configFiles[0]["permissions"])
	}
}

func TestExpandDeploymentDefinitionSetsDatabase(t *testing.T) {
	raw := testRawDefinition()
	raw["type"] = "DATABASE"
	raw["database"] = testSetOf(map[string]interface{}{
		"neon_postgres": testSetOf(map[string]interface{}{
			"pg_version":    16,
			"region":        "was",
			"instance_type": "free",
			"databases": testSetOf(map[string]interface{}{
				"name":  "koyebdb",
				"owner": "koyeb-adm",
			}),
			"roles": testSetOf(map[string]interface{}{
				"name":   "koyeb-adm",
				"secret": "role-secret",
			}),
		}),
	})

	definition := expandDeploymentDefinition(raw)

	if definition.Database == nil || definition.Database.NeonPostgres == nil {
		t.Fatal("expected a neon postgres database source to be set")
	}

	neon := definition.Database.NeonPostgres
	if neon.GetPgVersion() != 16 {
		t.Errorf("expected pg_version 16, got %d", neon.GetPgVersion())
	}
	if neon.GetRegion() != "was" {
		t.Errorf("expected region was, got %q", neon.GetRegion())
	}
	if neon.GetInstanceType() != "free" {
		t.Errorf("expected instance type free, got %q", neon.GetInstanceType())
	}
	if len(neon.Databases) != 1 || neon.Databases[0].GetName() != "koyebdb" || neon.Databases[0].GetOwner() != "koyeb-adm" {
		t.Errorf("expected one database koyebdb owned by koyeb-adm, got %+v", neon.Databases)
	}
	if len(neon.Roles) != 1 || neon.Roles[0].GetName() != "koyeb-adm" || neon.Roles[0].GetSecret() != "role-secret" {
		t.Errorf("expected one role koyeb-adm with a secret, got %+v", neon.Roles)
	}
}

func TestFlattenDeploymentDefinitionSetsDatabase(t *testing.T) {
	definition := &koyeb.DeploymentDefinition{
		Database: &koyeb.DatabaseSource{
			NeonPostgres: &koyeb.NeonPostgresDatabase{
				PgVersion:    toOpt(int64(16)),
				Region:       toOpt("was"),
				InstanceType: toOpt("free"),
				Databases: []koyeb.NeonPostgresDatabaseNeonDatabase{
					{Name: toOpt("koyebdb"), Owner: toOpt("koyeb-adm")},
				},
				Roles: []koyeb.NeonPostgresDatabaseNeonRole{
					{Name: toOpt("koyeb-adm"), Secret: toOpt("role-secret")},
				},
			},
		},
	}

	flattened := flattenDeploymentDefinition(definition)[0].(map[string]interface{})

	databases, ok := flattened["database"].([]interface{})
	if !ok || len(databases) != 1 {
		t.Fatalf("expected database to be flattened, got %T", flattened["database"])
	}
	database := databases[0].(map[string]interface{})
	neonPostgresSet, ok := database["neon_postgres"].(*schema.Set)
	if !ok {
		t.Fatalf("expected neon_postgres to be flattened, got %T", database["neon_postgres"])
	}
	neon := neonPostgresSet.List()[0].(map[string]interface{})

	if neon["pg_version"] != 16 {
		t.Errorf("expected pg_version 16, got %v", neon["pg_version"])
	}
	if neon["region"] != "was" {
		t.Errorf("expected region was, got %v", neon["region"])
	}
	if neon["instance_type"] != "free" {
		t.Errorf("expected instance type free, got %v", neon["instance_type"])
	}

	neonDatabases := neon["databases"].(*schema.Set).List()
	if len(neonDatabases) != 1 {
		t.Fatalf("expected 1 database, got %d", len(neonDatabases))
	}
	if neonDatabases[0].(map[string]interface{})["name"] != "koyebdb" {
		t.Errorf("expected database name koyebdb, got %v", neonDatabases[0])
	}
	if neonDatabases[0].(map[string]interface{})["owner"] != "koyeb-adm" {
		t.Errorf("expected database owner koyeb-adm, got %v", neonDatabases[0])
	}

	neonRoles := neon["roles"].(*schema.Set).List()
	if len(neonRoles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(neonRoles))
	}
	if neonRoles[0].(map[string]interface{})["name"] != "koyeb-adm" {
		t.Errorf("expected role name koyeb-adm, got %v", neonRoles[0])
	}
	if neonRoles[0].(map[string]interface{})["secret"] != "role-secret" {
		t.Errorf("expected role secret role-secret, got %v", neonRoles[0])
	}
}

func TestExpandDeploymentDefinitionSetsMesh(t *testing.T) {
	raw := testRawDefinition()
	raw["mesh"] = "DEPLOYMENT_MESH_ENABLED"

	definition := expandDeploymentDefinition(raw)

	if definition.GetMesh() != koyeb.DEPLOYMENTMESH_ENABLED {
		t.Errorf("expected mesh %q, got %q", koyeb.DEPLOYMENTMESH_ENABLED, definition.GetMesh())
	}
}

func TestFlattenDeploymentDefinitionSetsMesh(t *testing.T) {
	definition := &koyeb.DeploymentDefinition{
		Mesh: toOpt(koyeb.DEPLOYMENTMESH_ENABLED),
	}

	flattened := flattenDeploymentDefinition(definition)[0].(map[string]interface{})

	if flattened["mesh"] != string(koyeb.DEPLOYMENTMESH_ENABLED) {
		t.Errorf("expected mesh %q, got %v", koyeb.DEPLOYMENTMESH_ENABLED, flattened["mesh"])
	}
}

func TestExpandDeploymentDefinitionSetsNetworkPolicy(t *testing.T) {
	raw := testRawDefinition()
	raw["network_policy"] = testSetOf(map[string]interface{}{
		"egress": testSetOf(map[string]interface{}{
			"mode":       "EGRESS_POLICY_MODE_DENY_ALL",
			"allow_list": schema.NewSet(schema.HashString, []interface{}{"10.0.0.0/8"}),
		}),
		"mesh": testSetOf(map[string]interface{}{
			"scope": "MESH_SCOPE_APP",
			"name":  "",
		}),
	})

	definition := expandDeploymentDefinition(raw)

	if definition.NetworkPolicy == nil {
		t.Fatal("expected a network policy to be set")
	}

	if definition.NetworkPolicy.Egress == nil {
		t.Fatal("expected an egress policy to be set")
	}
	if definition.NetworkPolicy.Egress.GetMode() != koyeb.EGRESSPOLICYMODE_DENY_ALL {
		t.Errorf("expected egress mode %q, got %q", koyeb.EGRESSPOLICYMODE_DENY_ALL, definition.NetworkPolicy.Egress.GetMode())
	}
	if len(definition.NetworkPolicy.Egress.AllowList) != 1 || definition.NetworkPolicy.Egress.AllowList[0].GetCidr() != "10.0.0.0/8" {
		t.Errorf("expected allow list [10.0.0.0/8], got %+v", definition.NetworkPolicy.Egress.AllowList)
	}

	if definition.NetworkPolicy.Mesh == nil {
		t.Fatal("expected a mesh policy to be set")
	}
	if definition.NetworkPolicy.Mesh.GetScope() != koyeb.MESHSCOPE_APP {
		t.Errorf("expected mesh scope %q, got %q", koyeb.MESHSCOPE_APP, definition.NetworkPolicy.Mesh.GetScope())
	}
}

func TestFlattenDeploymentDefinitionSetsNetworkPolicy(t *testing.T) {
	definition := &koyeb.DeploymentDefinition{
		NetworkPolicy: &koyeb.NetworkPolicy{
			Egress: &koyeb.EgressPolicy{
				Mode: toOpt(koyeb.EGRESSPOLICYMODE_DENY_ALL),
				AllowList: []koyeb.NetworkPolicyDestination{
					{Cidr: toOpt("10.0.0.0/8")},
				},
			},
			Mesh: &koyeb.Mesh{
				Scope: toOpt(koyeb.MESHSCOPE_APP),
			},
		},
	}

	flattened := flattenDeploymentDefinition(definition)[0].(map[string]interface{})

	networkPolicies, ok := flattened["network_policy"].([]interface{})
	if !ok || len(networkPolicies) != 1 {
		t.Fatalf("expected network_policy to be flattened, got %T", flattened["network_policy"])
	}
	networkPolicy := networkPolicies[0].(map[string]interface{})

	egressSet, ok := networkPolicy["egress"].(*schema.Set)
	if !ok {
		t.Fatalf("expected egress to be flattened, got %T", networkPolicy["egress"])
	}
	egress := egressSet.List()[0].(map[string]interface{})
	if egress["mode"] != "EGRESS_POLICY_MODE_DENY_ALL" {
		t.Errorf("expected egress mode EGRESS_POLICY_MODE_DENY_ALL, got %v", egress["mode"])
	}
	allowList := egress["allow_list"].(*schema.Set).List()
	if len(allowList) != 1 || allowList[0] != "10.0.0.0/8" {
		t.Errorf("expected allow list [10.0.0.0/8], got %v", allowList)
	}

	meshSet, ok := networkPolicy["mesh"].(*schema.Set)
	if !ok {
		t.Fatalf("expected mesh to be flattened, got %T", networkPolicy["mesh"])
	}
	mesh := meshSet.List()[0].(map[string]interface{})
	if mesh["scope"] != "MESH_SCOPE_APP" {
		t.Errorf("expected mesh scope MESH_SCOPE_APP, got %v", mesh["scope"])
	}
}

func TestExpandRoutesSetsSecurityPolicies(t *testing.T) {
	config := []interface{}{
		map[string]interface{}{
			"port": 3000,
			"path": "/",
			"security_policies": testSetOf(map[string]interface{}{
				"basic_auths": testSetOf(map[string]interface{}{
					"username": "user",
					"password": "password",
				}),
				"api_keys": schema.NewSet(schema.HashString, []interface{}{"api-key"}),
			}),
		},
	}

	routes := expandRoutes(config)

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	if routes[0].SecurityPolicies == nil {
		t.Fatal("expected security policies to be set")
	}
	basicAuths := routes[0].SecurityPolicies.BasicAuths
	if len(basicAuths) != 1 {
		t.Fatalf("expected 1 basic auth, got %d", len(basicAuths))
	}
	if basicAuths[0].GetUsername() != "user" {
		t.Errorf("expected username user, got %q", basicAuths[0].GetUsername())
	}
	if basicAuths[0].GetPassword() != "password" {
		t.Errorf("expected password to be set, got %q", basicAuths[0].GetPassword())
	}
	apiKeys := routes[0].SecurityPolicies.ApiKeys
	if len(apiKeys) != 1 || apiKeys[0] != "api-key" {
		t.Errorf("expected api keys [api-key], got %+v", apiKeys)
	}
}

func TestFlattenRoutesSetsSecurityPolicies(t *testing.T) {
	routes := []koyeb.DeploymentRoute{
		{
			Port: toOpt(int64(3000)),
			Path: toOpt("/"),
			SecurityPolicies: &koyeb.SecurityPolicies{
				BasicAuths: []koyeb.BasicAuthPolicy{
					{Username: toOpt("user"), Password: toOpt("password")},
				},
				ApiKeys: []string{"api-key"},
			},
		},
	}

	flattened := flattenRoutes(&routes)

	route := flattened[0]
	securityPoliciesSet, ok := route["security_policies"].(*schema.Set)
	if !ok {
		t.Fatalf("expected security_policies to be flattened, got %T", route["security_policies"])
	}
	securityPolicies := securityPoliciesSet.List()[0].(map[string]interface{})

	basicAuths := securityPolicies["basic_auths"].(*schema.Set).List()
	if len(basicAuths) != 1 {
		t.Fatalf("expected 1 basic auth, got %d", len(basicAuths))
	}
	basicAuth := basicAuths[0].(map[string]interface{})
	if basicAuth["username"] != "user" {
		t.Errorf("expected username user, got %v", basicAuth["username"])
	}
	if basicAuth["password"] != "password" {
		t.Errorf("expected password to be set, got %v", basicAuth["password"])
	}

	apiKeys := securityPolicies["api_keys"].(*schema.Set).List()
	if len(apiKeys) != 1 || apiKeys[0] != "api-key" {
		t.Errorf("expected api keys [api-key], got %v", apiKeys)
	}
}

func TestExpandDeploymentDefinitionSetsArchive(t *testing.T) {
	raw := testRawDefinition()
	raw["archive"] = testSetOf(map[string]interface{}{
		"id": "archive-id",
		"buildpack": testSetOf(map[string]interface{}{
			"build_command": "npm run build",
			"run_command":   "npm start",
			"privileged":    false,
		}),
	})

	definition := expandDeploymentDefinition(raw)

	if definition.Archive == nil {
		t.Fatal("expected an archive source to be set")
	}
	if definition.Archive.GetId() != "archive-id" {
		t.Errorf("expected archive id archive-id, got %q", definition.Archive.GetId())
	}
	if buildpack, ok := definition.Archive.GetBuildpackOk(); ok {
		if buildpack.GetBuildCommand() != "npm run build" {
			t.Errorf("expected build command npm run build, got %q", buildpack.GetBuildCommand())
		}
		if buildpack.GetRunCommand() != "npm start" {
			t.Errorf("expected run command npm start, got %q", buildpack.GetRunCommand())
		}
	} else {
		t.Error("expected a buildpack builder in the archive source")
	}
}

func TestFlattenDeploymentDefinitionSetsArchive(t *testing.T) {
	definition := &koyeb.DeploymentDefinition{
		Archive: &koyeb.ArchiveSource{
			Id: toOpt("archive-id"),
			Buildpack: &koyeb.BuildpackBuilder{
				BuildCommand: toOpt("npm run build"),
				RunCommand:   toOpt("npm start"),
			},
		},
	}

	flattened := flattenDeploymentDefinition(definition)[0].(map[string]interface{})

	archives, ok := flattened["archive"].([]interface{})
	if !ok || len(archives) != 1 {
		t.Fatalf("expected archive to be flattened, got %T", flattened["archive"])
	}
	archive := archives[0].(map[string]interface{})

	if archive["id"] != "archive-id" {
		t.Errorf("expected archive id archive-id, got %v", archive["id"])
	}
	buildpacks := archive["buildpack"].([]interface{})
	if len(buildpacks) != 1 {
		t.Fatalf("expected 1 buildpack, got %d", len(buildpacks))
	}
	buildpack := buildpacks[0].(map[string]interface{})
	if buildpack["build_command"] != "npm run build" {
		t.Errorf("expected build command npm run build, got %v", buildpack["build_command"])
	}
	if buildpack["run_command"] != "npm start" {
		t.Errorf("expected run command npm start, got %v", buildpack["run_command"])
	}
}

func TestDeploymentDefinitionTypeAllowsSandbox(t *testing.T) {
	definitionType := deploymentDefinitionSchema().Schema["type"]

	// Every other client sets SANDBOX on pool definitions; the name
	// stays Required, which is what the server demands for SANDBOX.
	if _, errs := definitionType.ValidateFunc("SANDBOX", "type"); len(errs) != 0 {
		t.Errorf("expected SANDBOX to validate, got %v", errs)
	}
	for _, accepted := range []string{"WEB", "WORKER", "DATABASE"} {
		if _, errs := definitionType.ValidateFunc(accepted, "type"); len(errs) != 0 {
			t.Errorf("expected %s to keep validating, got %v", accepted, errs)
		}
	}
	if _, errs := definitionType.ValidateFunc("BOGUS", "type"); len(errs) == 0 {
		t.Error("expected BOGUS to stay rejected")
	}
}

func TestProxyPortSchemaOnlyAllowsTCPProtocol(t *testing.T) {
	protocol := proxyPortSchema().Schema["protocol"]

	if _, errs := protocol.ValidateFunc("tcp", "protocol"); len(errs) != 0 {
		t.Errorf("expected tcp to validate, got %v", errs)
	}
	if _, errs := protocol.ValidateFunc("http", "protocol"); len(errs) == 0 {
		t.Error("expected http to be rejected: the API enum only allows tcp")
	}
}

func TestFlattenNetworkPolicyEmptyProducesNoPhantomElement(t *testing.T) {
	flattened := flattenNetworkPolicy(&koyeb.NetworkPolicy{})

	if len(flattened) != 0 {
		t.Errorf("expected no element for an empty network policy, got %v", flattened)
	}
}

func TestExpandProxyPortsOmitsEmptyProtocol(t *testing.T) {
	// The SDK zero-fills absent keys to "", so pin the production shape.
	config := []interface{}{
		map[string]interface{}{"port": 22, "protocol": ""},
	}

	proxyPorts := expandProxyPorts(config)

	if len(proxyPorts) != 1 {
		t.Fatalf("expected 1 proxy port, got %d", len(proxyPorts))
	}
	if proxyPorts[0].Protocol != nil {
		t.Errorf("expected protocol to stay unset when omitted, got %q", proxyPorts[0].GetProtocol())
	}
}

// testSetOf builds a *schema.Set from raw items for expand tests; the hash
// function is irrelevant because the expand funcs only read List().
func testSetOf(items ...interface{}) *schema.Set {
	return schema.NewSet(func(_ interface{}) int { return 0 }, items)
}

func TestExpandVolumesMapsSchemaScopeKey(t *testing.T) {
	config := []interface{}{
		map[string]interface{}{
			"id":            "vol-id",
			"path":          "/data",
			"replica_index": 0,
			"scope":         []interface{}{"was"},
		},
	}

	volumes := expandVolumes(config)

	if len(volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(volumes))
	}
	if volumes[0].GetId() != "vol-id" || volumes[0].GetPath() != "/data" {
		t.Errorf("expected id vol-id at /data, got %+v", volumes[0])
	}
	if scopes := volumes[0].GetScopes(); len(scopes) != 1 || scopes[0] != "was" {
		t.Errorf("expected scopes [was], got %v", scopes)
	}
}

func TestFlattenVolumesSetsScope(t *testing.T) {
	volumes := []koyeb.DeploymentVolume{
		{
			Id:     toOpt("vol-id"),
			Path:   toOpt("/data"),
			Scopes: []string{"was"},
		},
	}

	flattened := flattenVolumes(&volumes)

	scope, ok := flattened[0]["scope"].([]string)
	if !ok || len(scope) != 1 || scope[0] != "was" {
		t.Errorf("expected scope [was], got %v (%T)", flattened[0]["scope"], flattened[0]["scope"])
	}
}

func TestDeploymentDefinitionSchemaMatchesAPIDefaults(t *testing.T) {
	s := deploymentDefinitionSchema().Schema

	// The API fills strategy and mesh with these values in every stored
	// definition; without matching schema defaults the definition
	// read-back produces a perpetual diff on the service pool resource.
	if got := s["strategy"].Default; got != "DEPLOYMENT_STRATEGY_TYPE_ROLLING" {
		t.Errorf("expected strategy default DEPLOYMENT_STRATEGY_TYPE_ROLLING, got %v", got)
	}
	if got := s["mesh"].Default; got != "DEPLOYMENT_MESH_AUTO" {
		t.Errorf("expected mesh default DEPLOYMENT_MESH_AUTO, got %v", got)
	}
}

// serviceWaitTestServer serves a create/update reply and a GetService that
// reports STARTING on the first poll and the given status afterwards.
func serviceWaitTestServer(t *testing.T, finalStatus string) (*httptest.Server, *int32) {
	t.Helper()
	var gets int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /v1/services", "PUT /v1/services/" + testServiceUUID:
			_, _ = w.Write([]byte(`{"service":{"id":"` + testServiceUUID + `","name":"my-service"}}`))
		case "GET /v1/services/" + testServiceUUID:
			status := firstPollThen(&gets, "STARTING", finalStatus)
			_, _ = w.Write([]byte(`{"service":{"id":"` + testServiceUUID + `","name":"my-service","status":"` + status + `"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, &gets
}

const testServiceUUID = "123e4567-e89b-42d3-a456-426614174001"

func testServiceRawDefinition() map[string]interface{} {
	return map[string]interface{}{
		"name": "my-service",
		"docker": []interface{}{
			map[string]interface{}{"image": "koyeb/demo"},
		},
	}
}

// Both usable statuses end the wait: DEGRADED services are handed to
// users just like HEALTHY ones (the Python SDK's ready set).
func TestResourceKoyebServiceCreateWaitsForServiceHealth(t *testing.T) {
	for _, finalStatus := range []string{"HEALTHY", "DEGRADED"} {
		t.Run(finalStatus, func(t *testing.T) {
			shortenWaits(t)
			srv, gets := serviceWaitTestServer(t, finalStatus)
			defer srv.Close()

			cfg := koyeb.NewConfiguration()
			cfg.Servers[0].URL = srv.URL

			// A UUIDv4 app name short-circuits the app mapper, so the mock
			// only needs the service endpoints.
			d := schema.TestResourceDataRaw(t, serviceSchema(), map[string]interface{}{
				"app_name":   "123e4567-e89b-42d3-a456-426614174000",
				"definition": []interface{}{testServiceRawDefinition()},
			})

			diags := resourceKoyebServiceCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

			if len(diags) != 0 {
				t.Fatalf("expected no diagnostics, got %v", diags)
			}
			if got := d.Get("status").(string); got != finalStatus {
				t.Errorf("expected the create to return a %s service, got %q", finalStatus, got)
			}
			if n := atomic.LoadInt32(gets); n < 2 {
				t.Errorf("expected the create to poll GetService at least twice, got %d polls", n)
			}
		})
	}
}

func TestResourceKoyebServiceCreateFailsWhenServiceNeverHealthy(t *testing.T) {
	shortenWaits(t)
	srv, _ := serviceWaitTestServer(t, "STARTING")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, serviceSchema(), map[string]interface{}{
		"app_name":   "123e4567-e89b-42d3-a456-426614174000",
		"definition": []interface{}{testServiceRawDefinition()},
	})

	diags := resourceKoyebServiceCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "Error waiting for service") {
		t.Errorf("expected a wait-failure error, got: %s", diags[0].Summary)
	}
}

func TestResourceKoyebServiceUpdateWaitsForServiceHealth(t *testing.T) {
	shortenWaits(t)
	srv, gets := serviceWaitTestServer(t, "HEALTHY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	state := &terraform.InstanceState{
		ID: testServiceUUID,
		Attributes: map[string]string{
			"app_name":                    "my-app",
			"definition.#":                "1",
			"definition.0.name":           "my-service",
			"definition.0.docker.#":       "1",
			"definition.0.docker.0.image": "koyeb/demo",
		},
	}
	d := resourceKoyebService().Data(state)

	diags := resourceKoyebServiceUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := d.Get("status").(string); got != "HEALTHY" {
		t.Errorf("expected the update to return a HEALTHY service, got %q", got)
	}
	if n := atomic.LoadInt32(gets); n < 2 {
		t.Errorf("expected the update to poll GetService at least twice, got %d polls", n)
	}
}

func TestResourceKoyebServiceCreateFailsFastWhenServiceUnhealthy(t *testing.T) {
	shortenWaits(t)
	srv, gets := serviceWaitTestServer(t, "UNHEALTHY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := schema.TestResourceDataRaw(t, serviceSchema(), map[string]interface{}{
		"app_name":   "123e4567-e89b-42d3-a456-426614174000",
		"definition": []interface{}{testServiceRawDefinition()},
	})

	diags := resourceKoyebServiceCreate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "UNHEALTHY") {
		t.Errorf("expected the diagnostic to surface the UNHEALTHY service status, got: %s", diags[0].Summary)
	}
	if n := atomic.LoadInt32(gets); n != 2 {
		t.Errorf("expected the wait to stop at the first UNHEALTHY poll, got %d polls", n)
	}
}

// deploymentWaitTestServer serves a service update whose reply pins the
// replacement deployment, a service GET that keeps reporting the OLD
// deployment's HEALTHY, and a deployment that reports STARTING on the
// first poll and finalStatus afterwards.
func deploymentWaitTestServer(t *testing.T, pinsReply bool, finalStatus string) (*httptest.Server, *int32, *int32) {
	t.Helper()
	var serviceGets, deploymentGets int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "PUT /v1/services/" + testServiceUUID:
			latest := ""
			if pinsReply {
				latest = `,"latest_deployment_id":"dep-uuid"`
			}
			_, _ = w.Write([]byte(`{"service":{"id":"` + testServiceUUID + `","name":"my-service"` + latest + `}}`))
		case "GET /v1/services/" + testServiceUUID:
			atomic.AddInt32(&serviceGets, 1)
			// The old deployment stays healthy through the rollout: the
			// service-level status must never satisfy the update wait.
			_, _ = w.Write([]byte(`{"service":{"id":"` + testServiceUUID + `","name":"my-service",` +
				`"status":"HEALTHY","latest_deployment_id":"dep-uuid"}}`))
		case "GET /v1/deployments/dep-uuid":
			status := firstPollThen(&deploymentGets, "STARTING", finalStatus)
			_, _ = w.Write([]byte(`{"deployment":{"id":"dep-uuid","status":"` + status + `",` +
				`"messages":["build failed: could not read source"]}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, &serviceGets, &deploymentGets
}

func testServiceUpdateState() *terraform.InstanceState {
	return &terraform.InstanceState{
		ID: testServiceUUID,
		Attributes: map[string]string{
			"app_name":                    "my-app",
			"definition.#":                "1",
			"definition.0.name":           "my-service",
			"definition.0.docker.#":       "1",
			"definition.0.docker.0.image": "koyeb/demo",
		},
	}
}

// An update must be verified against the REPLACEMENT deployment, not the
// predecessor: the old deployment keeps the service-level status HEALTHY
// through a rolling update, so only the pinned deployment's transition to
// healthy proves the new version is live (mirrors the Python SDK fix
// koyeb-python-sdk@10210e3).
func TestResourceKoyebServiceUpdateWaitsForReplacementDeployment(t *testing.T) {
	shortenWaits(t)
	srv, _, deploymentGets := deploymentWaitTestServer(t, true, "HEALTHY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := resourceKoyebService().Data(testServiceUpdateState())

	diags := resourceKoyebServiceUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if n := atomic.LoadInt32(deploymentGets); n < 2 {
		t.Errorf("expected the update to poll GetDeployment at least twice, got %d polls", n)
	}
}

// The update reply does not always carry the replacement id; the live
// service does, and pinning from it keeps the wait on the replacement.
func TestResourceKoyebServiceUpdatePinsDeploymentFromService(t *testing.T) {
	shortenWaits(t)
	srv, serviceGets, deploymentGets := deploymentWaitTestServer(t, false, "HEALTHY")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := resourceKoyebService().Data(testServiceUpdateState())

	diags := resourceKoyebServiceUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if n := atomic.LoadInt32(serviceGets); n < 1 {
		t.Errorf("expected the pin to be read from the live service, got %d service polls", n)
	}
	if n := atomic.LoadInt32(deploymentGets); n < 2 {
		t.Errorf("expected the update to poll GetDeployment at least twice, got %d polls", n)
	}
}

// A replacement deployment that errors is terminal: the update fails fast
// with the status instead of burning the whole budget.
func TestResourceKoyebServiceUpdateFailsWhenReplacementErrors(t *testing.T) {
	shortenWaits(t)
	srv, _, deploymentGets := deploymentWaitTestServer(t, true, "ERROR")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := resourceKoyebService().Data(testServiceUpdateState())

	diags := resourceKoyebServiceUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "ERROR") {
		t.Errorf("expected the diagnostic to surface the ERROR deployment status, got: %s", diags[0].Summary)
	}
	if n := atomic.LoadInt32(deploymentGets); n != 2 {
		t.Errorf("expected the wait to stop at the first ERROR poll, got %d polls", n)
	}
}

// The deployment's own messages explain why it failed; the wait surfaces
// them instead of a bare status.
func TestResourceKoyebServiceUpdateSurfacesDeploymentMessages(t *testing.T) {
	shortenWaits(t)
	srv, _, _ := deploymentWaitTestServer(t, true, "ERROR")
	defer srv.Close()

	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL

	d := resourceKoyebService().Data(testServiceUpdateState())

	diags := resourceKoyebServiceUpdate(context.Background(), d, koyeb.NewAPIClient(cfg))

	if len(diags) != 1 || diags[0].Severity != diag.Error {
		t.Fatalf("expected exactly 1 error diagnostic, got %v", diags)
	}
	if !strings.Contains(diags[0].Summary, "build failed: could not read source") {
		t.Errorf("expected the diagnostic to surface the deployment messages, got: %s", diags[0].Summary)
	}
}
