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
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

const testNamePrefix = "tf-acc-test-"

var testAccProvider *schema.Provider
var testAccProviders map[string]*schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func init() {
	testAccProvider = New("test")()
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

func TestProviderInternalValidate(t *testing.T) {
	if err := testAccProvider.InternalValidate(); err != nil {
		t.Fatalf("provider InternalValidate: %s", err)
	}
}

func testAccSkipIfServicePoolsUnavailable(t *testing.T) {
	// Service pools ship behind the serverless_container_sandbox_pool
	// flag and a route deployment: APIs that don't serve them yet cannot
	// pass these tests, so probe before running.
	client := testAccProvider.Meta().(*koyeb.APIClient)
	if _, _, err := client.ServicePoolsApi.ListServicePools(context.Background()).Execute(); err != nil {
		t.Skipf("service pools are not available on the target Koyeb API: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("KOYEB_TOKEN"); v == "" {
		// Fork PRs get no secrets, so CI cannot run ACC tests there;
		// skipping keeps the suite meaningful without credentials.
		t.Skip("KOYEB_TOKEN must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}
