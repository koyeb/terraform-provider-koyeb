package koyeb

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedConfig() (interface{}, error) {
	if os.Getenv("KOYEB_TOKEN") == "" {
		return nil, fmt.Errorf("Empty KOYEB_TOKEN environment variable")
	}

	p := schema.Provider{}
	userAgent := p.UserAgent("terraform-provider-koyeb", "test")

	apiHost := os.Getenv("KOYEB_API_URL")
	if apiHost == "" {
		apiHost = "app.koyeb.com"
	}

	koyebClientConfig := koyeb.NewConfiguration()
	koyebClientConfig.DefaultHeader["Authorization"] = "Bearer " + os.Getenv("KOYEB_TOKEN")
	koyebClientConfig.UserAgent = userAgent

	client := koyeb.NewAPIClient(koyebClientConfig)

	return client, nil
}
