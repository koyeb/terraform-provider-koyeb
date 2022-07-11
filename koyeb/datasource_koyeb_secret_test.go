package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccDataSourceKoyebSecret_Basic(t *testing.T) {
	var secret koyeb.Secret
	secretName := randomTestName()
	secretValue := randomTestName()

	resourceConfig := fmt.Sprintf(`
resource "koyeb_secret" "foo" {
  name       = "%s"
  value      = "%s"
}`, secretName, secretValue)

	dataSourceConfig := `
data "koyeb_secret" "bar" {
  name = koyeb_secret.foo.name
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: resourceConfig,
			},
			{
				Config: resourceConfig + dataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceKoyebSecretExists("data.koyeb_secret.bar", &secret),
					testAccCheckDataSourceKoyebSecretAttributes(&secret, secretName),
					resource.TestCheckResourceAttr(
						"data.koyeb_secret.bar", "name", secretName),
					resource.TestCheckResourceAttrSet("data.koyeb_secret.bar", "id"),
					resource.TestCheckResourceAttrSet("data.koyeb_secret.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_secret.bar", "type"),
				),
			},
		},
	})
}

func testAccCheckDataSourceKoyebSecretAttributes(secret *koyeb.Secret, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *secret.Name != name {
			return fmt.Errorf("Bad name: %s", *secret.Name)
		}

		return nil
	}
}

func testAccCheckDataSourceKoyebSecretExists(n string, secret *koyeb.Secret) resource.TestCheckFunc {
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
