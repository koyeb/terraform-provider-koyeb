package koyeb

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			DataSourcesMap: map[string]*schema.Resource{
				"koyeb_app":     dataSourceKoyebApp(),
				"koyeb_service": dataSourceKoyebService(),
				"koyeb_domain":  dataSourceKoyebDomain(),
				"koyeb_secret":  dataSourceKoyebSecret(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"koyeb_app":     resourceKoyebApp(),
				"koyeb_service": resourceKoyebService(),
				"koyeb_domain":  resourceKoyebDomain(),
				"koyeb_secret":  resourceKoyebSecret(),
				"koyeb_volume":  resourceKoyebVolume(),
			},
		}

		p.ConfigureContextFunc = configure(p, version)

		return p
	}
}

func configure(p *schema.Provider, version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
		if os.Getenv("KOYEB_TOKEN") == "" {
			return nil, diag.Errorf("Empty KOYEB_TOKEN environment variable")
		}

		userAgent := p.UserAgent("terraform-provider-koyeb", version)
		koyebClientConfig := koyeb.NewConfiguration()
		koyebClientConfig.Host = "app.koyeb.com"
		koyebClientConfig.Debug = os.Getenv("KOYEB_DEBUG") == "true"
		koyebClientConfig.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", os.Getenv("KOYEB_TOKEN"))
		koyebClientConfig.UserAgent = userAgent

		client := koyeb.NewAPIClient(koyebClientConfig)
		return client, nil
	}
}
