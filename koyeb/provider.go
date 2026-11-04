package koyeb

import (
	"context"
	"fmt"
	"net/url"
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
			Schema: map[string]*schema.Schema{
				"api_url": {
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("KOYEB_API_URL", defaultAPIURL),
					Description:  "The URL of the Koyeb API. Useful to point the provider at a staging or local API.",
					ValidateFunc: validateAPIURL,
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"koyeb_app":          dataSourceKoyebApp(),
				"koyeb_service":      dataSourceKoyebService(),
				"koyeb_domain":       dataSourceKoyebDomain(),
				"koyeb_secret":       dataSourceKoyebSecret(),
				"koyeb_service_pool": dataSourceKoyebServicePool(),
				"koyeb_snapshot":     dataSourceKoyebSnapshot(),
				"koyeb_database":     dataSourceKoyebDatabase(),
				"koyeb_volume":       dataSourceKoyebVolume(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"koyeb_app":                resourceKoyebApp(),
				"koyeb_service":            resourceKoyebService(),
				"koyeb_database":           resourceKoyebDatabase(),
				"koyeb_domain":             resourceKoyebDomain(),
				"koyeb_secret":             resourceKoyebSecret(),
				"koyeb_volume":             resourceKoyebVolume(),
				"koyeb_snapshot":           resourceKoyebSnapshot(),
				"koyeb_service_pool":       resourceKoyebServicePool(),
				"koyeb_service_pool_claim": resourceKoyebServicePoolClaim(),
			},
		}

		p.ConfigureContextFunc = configure(p, version)

		return p
	}
}

func configure(p *schema.Provider, version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		if os.Getenv("KOYEB_TOKEN") == "" {
			return nil, diag.Errorf("Empty KOYEB_TOKEN environment variable")
		}

		userAgent := p.UserAgent("terraform-provider-koyeb", version)
		koyebClientConfig := koyeb.NewConfiguration()
		if err := applyAPIURL(koyebClientConfig, d.Get("api_url").(string)); err != nil {
			return nil, diag.FromErr(err)
		}
		koyebClientConfig.Debug = os.Getenv("KOYEB_DEBUG") == "true"
		koyebClientConfig.DefaultHeader["Authorization"] = fmt.Sprintf("Bearer %s", os.Getenv("KOYEB_TOKEN"))
		koyebClientConfig.UserAgent = userAgent

		client := koyeb.NewAPIClient(koyebClientConfig)
		return client, nil
	}
}

const defaultAPIURL = "https://app.koyeb.com"

func validateAPIURL(v interface{}, k string) (ws []string, errors []error) {
	if _, err := parseAPIURL(v.(string)); err != nil {
		errors = append(errors, fmt.Errorf("%q is not a valid api_url: %w", k, err))
	}
	return
}

func parseAPIURL(apiURL string) (*url.URL, error) {
	u, err := url.Parse(apiURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("expected an http(s) URL with a host, got %q", apiURL)
	}
	return u, nil
}

// applyAPIURL points the client at apiURL; the generated client overrides
// the server URL's host and scheme from cfg.Host/cfg.Scheme.
func applyAPIURL(cfg *koyeb.Configuration, apiURL string) error {
	u, err := parseAPIURL(apiURL)
	if err != nil {
		return err
	}
	cfg.Host = u.Host
	cfg.Scheme = u.Scheme
	return nil
}
