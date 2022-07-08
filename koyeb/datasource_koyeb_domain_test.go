package koyeb

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestAccDataSourceKoyebDomain_Basic(t *testing.T) {
	var domain koyeb.Domain
	domainName := randomTestName() + ".com"

	resourceConfig := fmt.Sprintf(`
resource "koyeb_domain" "foo" {
  name = "%s"
}
`, domainName)

	dataSourceConfig := `
data "koyeb_domain" "bar" {
  name = koyeb_domain.foo.name
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
					testAccCheckDataSourceKoyebDomainExists("data.koyeb_domain.bar", &domain),
					testAccCheckDataSourceKoyebDomainAttributes(&domain, domainName),
					resource.TestCheckResourceAttr(
						"data.koyeb_domain.bar", "name", domainName),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "id"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "organization_id"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "type"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "intended_cname"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "status"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "messages"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "version"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "verified_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "updated_at"),
					resource.TestCheckResourceAttrSet("data.koyeb_domain.bar", "created_at"),
				),
			},
		},
	})
}

func testAccCheckDataSourceKoyebDomainAttributes(domain *koyeb.Domain, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *domain.Name != name {
			return fmt.Errorf("Bad name: %s", *domain.Name)
		}

		return nil
	}
}

func testAccCheckDataSourceKoyebDomainExists(n string, domain *koyeb.Domain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Record ID is set")
		}

		client := testAccProvider.Meta().(*koyeb.APIClient)

		res, _, err := client.DomainsApi.GetDomain(context.Background(), rs.Primary.ID).Execute()

		if err != nil {
			return err
		}

		if *res.Domain.Id != rs.Primary.ID {
			return fmt.Errorf("Record not found")
		}

		d := res.GetDomain()
		*domain = d

		return nil
	}
}
