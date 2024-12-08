package koyeb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func init() {
	resource.AddTestSweepers("koyeb_secret", &resource.Sweeper{
		Name: "koyeb_secret",
		F:    testSweepSecret,
	})
}

func testSweepSecret(region string) error {
	meta, err := sharedConfig()
	if err != nil {
		return fmt.Errorf("error retrieving shared config: %w", err)
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.SecretsApi.ListSecrets(context.Background()).Limit("100").Execute()
	if err != nil {
		return fmt.Errorf("error listing secrets: %w", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(res.Secrets))

	for _, secret := range res.Secrets {
		if !strings.HasPrefix(secret.GetName(), testNamePrefix) {
			continue
		}

		wg.Add(1)
		go func(secretID, secretName string) {
			defer wg.Done()
			log.Printf("[INFO] Destroying secret: %s", secretName)
			if _, _, err := client.SecretsApi.DeleteSecret(context.Background(), secretID).Execute(); err != nil {
				errs <- fmt.Errorf("error deleting secret %s: %w", secretName, err)
			}
		}(secret.GetId(), *secret.Name)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func TestAccKoyebSecret_Basic(t *testing.T) {
	var secret koyeb.Secret

	configs := []struct {
		nameSuffix   string
		templateType string
		extraArgs    string
	}{
		{"basic", "basic", ""},
		{"docker", "docker_hub_registry", randomTestName()},
		{"github", "github_registry", randomTestName()},
		{"gitlab", "gitlab_registry", randomTestName()},
		{"digitalocean", "digital_ocean_container_registry", randomTestName()},
		{"private", "private_registry", randomTestName()},
		// Uncomment when ready for Azure testing
		// {"azure", "azure_container_registry", randomTestName()},
	}

	for _, cfg := range configs {
		t.Run(fmt.Sprintf("Testing %s", cfg.nameSuffix), func(t *testing.T) {
			secretName := randomTestName() + "_" + cfg.nameSuffix
			secretValue := randomTestName()

			resource.ParallelTest(t, resource.TestCase{
				PreCheck:          func() { testAccPreCheck(t) },
				ProviderFactories: testAccProviderFactories,
				CheckDestroy:      testAccCheckKoyebSecretDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccKoyebSecretConfig(cfg.templateType, secretName, secretValue, cfg.extraArgs),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckKoyebSecretExists("koyeb_secret.foo", &secret),
							testAccCheckKoyebSecretAttributesMatch("koyeb_secret.foo", secretName),
						),
					},
				},
			})
		})
	}
}

func testAccCheckKoyebSecretDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_secret" {
			continue
		}

		_, _, err := client.SecretsApi.GetSecret(context.Background(), rs.Primary.ID).Execute()
		if err == nil {
			return fmt.Errorf("Secret still exists: %s ", err)
		}
	}

	return nil
}

func testAccCheckKoyebSecretAttributesMatch(resourceName, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)
		secret, _, err := client.SecretsApi.GetSecret(context.Background(), rs.Primary.ID).Execute()
		if err != nil {
			return fmt.Errorf("error retrieving secret: %w", err)
		}

		if *secret.Secret.Name != expectedName {
			return fmt.Errorf("expected name %s, got %s", expectedName, *secret.Secret.Name)
		}
		return nil
	}
}

func testAccCheckKoyebSecretExists(n string, secret *koyeb.Secret) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.SecretsApi.GetSecret(context.Background(), rs.Primary.ID).Execute()

		if err != nil {
			return err
		}

		if *res.Secret.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		*secret = res.GetSecret()

		return nil
	}
}

func testAccKoyebSecretConfig(templateType, name, value, extraArgs string) string {
	switch templateType {
	case "basic":
		return fmt.Sprintf(`
			resource "koyeb_secret" "foo" {
				name  = "%s"
				value = "%s"
			}`, name, value)
	case "docker_hub_registry":
		return fmt.Sprintf(`
			resource "koyeb_secret" "foo" {
				name  = "%s"
				type  = "REGISTRY"
				docker_hub_registry {
					username = "%s"
					password = "%s"
				}
			}`, name, value, extraArgs)
	case "github_registry":
		return fmt.Sprintf(`
			resource "koyeb_secret" "foo" {
				name  = "%s"
				type  = "REGISTRY"
				github_registry {
					username = "%s"
					password = "%s"
				}
			}`, name, value, extraArgs)
	case "gitlab_registry":
		return fmt.Sprintf(`
			resource "koyeb_secret" "foo" {
				name  = "%s"
				type  = "REGISTRY"
				gitlab_registry {
					username = "%s"
					password = "%s"
				}
			}`, name, value, extraArgs)
	case "digital_ocean_container_registry":
		return fmt.Sprintf(`
			resource "koyeb_secret" "foo" {
				name  = "%s"
				type  = "REGISTRY"
				digital_ocean_container_registry {
					username = "%s"
					password = "%s"
				}
			}`, name, value, extraArgs)
	case "private_registry":
		return fmt.Sprintf(`
			resource "koyeb_secret" "foo" {
				name  = "%s"
				type  = "REGISTRY"
				private_registry {
					username = "%s"
					password = "%s"
					url      = "%s"
				}
			}`, name, value, value, extraArgs)
	default:
		return ""
	}
}
