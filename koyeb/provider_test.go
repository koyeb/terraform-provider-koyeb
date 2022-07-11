package koyeb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testNamePrefix = "tf-acc-test-"

var testAccProvider *schema.Provider
var testAccProviders map[string]*schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProvider = New()()
	testAccProviders = map[string]*schema.Provider{
		"koyeb": testAccProvider,
	}
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"koyeb": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func randomTestName(additionalNames ...string) string {
	prefix := testNamePrefix
	for _, n := range additionalNames {
		prefix += "-" + strings.Replace(n, " ", "_", -1)
	}
	return randomName(prefix, 10)
}

func randomName(prefix string, length int) string {
	return fmt.Sprintf("%s%s", prefix, acctest.RandString(length))
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("KOYEB_TOKEN"); v == "" {
		t.Fatal("KOYEB_TOKEN must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}
