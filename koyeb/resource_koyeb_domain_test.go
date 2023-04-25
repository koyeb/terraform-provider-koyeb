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
	resource.AddTestSweepers("koyeb_domain", &resource.Sweeper{
		Name:         "koyeb_domain",
		F:            testSweepDomain,
		Dependencies: []string{"koyeb_app"},
	})

}

func testSweepDomain(string) error {
	meta, err := sharedConfig()
	if err != nil {
		return err
	}

	client := meta.(*koyeb.APIClient)

	res, _, err := client.DomainsApi.ListDomains(context.Background()).Limit("100").Execute()
	if err != nil {
		return err
	}

	for _, d := range res.Domains {
		if strings.HasPrefix(d.GetName(), testNamePrefix) {
			log.Printf("Destroying domain %s", *d.Name)

			if _, _, err := client.DomainsApi.DeleteDomain(context.Background(), d.GetId()).Execute(); err != nil {
				return err
			}
		}
	}

	return nil
}

func TestAccKoyebDomain_Basic(t *testing.T) {
	var domain koyeb.Domain
	appName := randomTestName()
	domainName := appName + ".com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckKoyebDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCheckKoyebDomainConfig_basic, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebDomainExists("koyeb_domain.foo", &domain),
					testAccCheckKoyebDomainAttributes(&domain, domainName),
					resource.TestCheckResourceAttr(
						"koyeb_domain.foo", "name", domainName),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "id"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "type"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "intended_cname"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "status"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "messages"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "version"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "verified_at"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "created_at"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCheckKoyebDomainConfig_basic_app_name_update, appName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKoyebDomainExists("koyeb_domain.foo", &domain),
					testAccCheckKoyebDomainAttributes(&domain, domainName),
					resource.TestCheckResourceAttr(
						"koyeb_domain.foo", "name", domainName),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "id"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "organization_id"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "type"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "intended_cname"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "status"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "messages"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "version"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "verified_at"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "updated_at"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "created_at"),
					resource.TestCheckResourceAttrSet("koyeb_domain.foo", "app_name"),
				),
			},
		},
	})
}

func testAccCheckKoyebDomainDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*koyeb.APIClient)
	targetStatus := []string{"DELETED", "DELETING"}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "koyeb_domain" {
			continue
		}

		err := waitForResourceStatus(client.DomainsApi.GetDomain(context.Background(), rs.Primary.ID).Execute, "Domain", targetStatus, 1, false)
		if err != nil {
			return fmt.Errorf("Domain still exists: %s ", err)
		}
	}

	return nil
}

func testAccCheckKoyebDomainAttributes(domain *koyeb.Domain, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *domain.Name != name {
			return fmt.Errorf("Bad name: %s", *domain.Name)
		}

		return nil
	}
}

func testAccCheckKoyebDomainExists(n string, domain *koyeb.Domain) resource.TestCheckFunc {
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

		*domain = res.GetDomain()

		return nil
	}
}

const testAccCheckKoyebDomainConfig_basic = `
resource "koyeb_domain" "foo" {
	name       = "%s"
}`

const testAccCheckKoyebDomainConfig_basic_app_name_update = `
resource "koyeb_app" "bar" {
	name = "%s"
}

resource "koyeb_domain" "foo" {
	name       = "%s"
	app_name   = "${koyeb_app.bar.name}"
}`
