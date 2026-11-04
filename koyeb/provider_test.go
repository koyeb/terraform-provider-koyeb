package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
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

func TestApplyAPIURL(t *testing.T) {
	tests := []struct {
		name    string
		apiURL  string
		wantErr bool
		host    string
		scheme  string
	}{
		{"default prod", "https://app.koyeb.com", false, "app.koyeb.com", "https"},
		{"http with port", "http://localhost:8080", false, "localhost:8080", "http"},
		{"custom https host", "https://staging.koyeb.example", false, "staging.koyeb.example", "https"},
		{"missing scheme", "app.koyeb.com", true, "", ""},
		{"missing host", "https://", true, "", ""},
		{"garbage", "://nope", true, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := koyeb.NewConfiguration()
			err := applyAPIURL(cfg, tt.apiURL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %q", tt.apiURL)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.apiURL, err)
			}
			if cfg.Host != tt.host || cfg.Scheme != tt.scheme {
				t.Errorf("got host=%q scheme=%q, want host=%q scheme=%q", cfg.Host, cfg.Scheme, tt.host, tt.scheme)
			}
		})
	}
}

// The configured client must actually call the API URL from the provider
// schema, not the hardcoded default.
func TestProviderConfigureHonorsAPIURLAttribute(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apps":[],"count":0}`))
	}))
	defer srv.Close()

	t.Setenv("KOYEB_TOKEN", "test-token")

	p := New("test")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
		"api_url": srv.URL,
	})
	meta, diags := p.ConfigureContextFunc(context.Background(), d)
	if diags.HasError() {
		t.Fatalf("configure: %v", diags)
	}

	client := meta.(*koyeb.APIClient)
	if _, _, err := client.AppsApi.ListApps(context.Background()).Limit("10").Offset("0").Execute(); err != nil {
		t.Fatalf("ListApps against the custom API URL failed: %v", err)
	}
	if n := atomic.LoadInt32(&hits); n == 0 {
		t.Fatal("expected the client to call the configured api_url, got no request")
	}
}

// KOYEB_API_URL must work without any provider block attribute.
func TestProviderConfigureHonorsAPIURLEnv(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"apps":[],"count":0}`))
	}))
	defer srv.Close()

	t.Setenv("KOYEB_TOKEN", "test-token")
	t.Setenv("KOYEB_API_URL", srv.URL)

	p := New("test")()
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{})
	meta, diags := p.ConfigureContextFunc(context.Background(), d)
	if diags.HasError() {
		t.Fatalf("configure: %v", diags)
	}

	client := meta.(*koyeb.APIClient)
	if _, _, err := client.AppsApi.ListApps(context.Background()).Limit("10").Offset("0").Execute(); err != nil {
		t.Fatalf("ListApps against the env API URL failed: %v", err)
	}
	if n := atomic.LoadInt32(&hits); n == 0 {
		t.Fatal("expected the client to call KOYEB_API_URL, got no request")
	}
}

func TestValidateAPIURL(t *testing.T) {
	if _, errs := validateAPIURL("https://app.koyeb.com", "api_url"); len(errs) != 0 {
		t.Fatalf("valid api_url rejected: %v", errs)
	}
	for _, bad := range []string{"app.koyeb.com", "https://", "://nope"} {
		if _, errs := validateAPIURL(bad, "api_url"); len(errs) == 0 {
			t.Errorf("api_url %q without scheme+host must be rejected", bad)
		}
	}
}

func TestParseAPIURLRejectsPath(t *testing.T) {
	if _, err := parseAPIURL("https://app.koyeb.com/v2"); err == nil {
		t.Fatal("api_url with a path prefix must be rejected, not silently dropped")
	}
}
