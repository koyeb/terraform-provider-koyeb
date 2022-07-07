package koyeb

import (
	"context"
	"fmt"
	"log"
	"strings"
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

func testSweepSecret(string) error {
	meta, err := sharedConfig()
	if err != nil {
		return err
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.SecretsApi.ListSecrets(context.Background()).Limit("100").Execute()
	if err != nil {
		return err
	}

	for _, d := range *res.Secrets {
		if strings.HasPrefix(d.GetName(), testNamePrefix) {
			log.Printf("Destroying secret %s", *d.Name)

			if _, _, err := client.SecretsApi.DeleteSecret(context.Background(), d.GetId()).Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccKoyebSecret_Basic(t *testing.T) {
	var secret koyeb.Secret
	secretName := randomTestName()
	secretValue := randomTestName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebSecretDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebSecretConfig_basic, secretName, secretValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebSecretExists("koyeb_secret.foo", &secret),
					testAccCheckKoyebSecretAttributes(&secret, secretName),
					resource.TestCheckResourceAttr(
						"koyeb_secret.foo", "name", secretName),
					resource.TestCheckResourceAttrSet("koyeb_secret.foo", "id"),
					resource.TestCheckResourceAttrSet("koyeb_secret.foo", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_secret.foo", "type"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebSecretConfig_basic_type_update, secretName, secretValue, secretValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebSecretExists("koyeb_secret.foo", &secret),
					testAccCheckKoyebSecretAttributes(&secret, secretName),
					resource.TestCheckResourceAttr(
						"koyeb_secret.foo", "name", secretName),
					resource.TestCheckResourceAttrSet("koyeb_secret.foo", "id"),
					resource.TestCheckResourceAttrSet("koyeb_secret.foo", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_secret.foo", "type"),
				),
			},
		},
	})
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

func testAccCheckKoyebSecretAttributes(secret *koyeb.Secret, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *secret.Name != name {
			return fmt.Errorf("Bad name: %s", *secret.Name)
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

		d := res.GetSecret()
		*secret = d

		return nil
	}
}

const testAccCheckKoyebSecretConfig_basic = `
resource "koyeb_secret" "foo" {
	name       = "%s"
	value      = "%s"
}`

const testAccCheckKoyebSecretConfig_basic_type_update = `
resource "koyeb_secret" "foo" {
	name  = "%s"
	type  = "REGISTRY"
	docker_hub_registry {
		username = "%s"
		password = "%s"
	}
}`
