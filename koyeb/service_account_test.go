package koyeb

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// The service account is a service-level attribute the API accepts at
// creation only (CreateService.service_account_id; UpdateService has no such
// field), so the resource exposes it as ForceNew and only the create payload
// carries it.
func TestServiceSchemaExposesServiceAccountID(t *testing.T) {
	attr, ok := serviceSchema()["service_account_id"]
	if !ok {
		t.Fatal("service_account_id missing from service schema")
	}
	if attr.Type != schema.TypeString {
		t.Fatalf("service_account_id type = %v, want TypeString", attr.Type)
	}
	if !attr.Optional || !attr.Computed {
		t.Fatalf("service_account_id must be Optional+Computed, got Optional=%v Computed=%v", attr.Optional, attr.Computed)
	}
	if !attr.ForceNew {
		t.Fatal("service_account_id must be ForceNew: the API rejects changing it after creation")
	}
}

func TestCreateServicePayloadCarriesServiceAccountID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, serviceSchema(), map[string]interface{}{
		"app_name":           "app",
		"service_account_id": "sa-123",
		"definition":         []interface{}{map[string]interface{}{}},
	})
	payload := createServicePayload("app-id", d)
	if payload.ServiceAccountId == nil || *payload.ServiceAccountId != "sa-123" {
		t.Fatalf("create payload ServiceAccountId = %v, want sa-123", payload.ServiceAccountId)
	}
}

func TestCreateServicePayloadOmitsEmptyServiceAccountID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, serviceSchema(), map[string]interface{}{
		"app_name":   "app",
		"definition": []interface{}{map[string]interface{}{}},
	})
	payload := createServicePayload("app-id", d)
	if payload.ServiceAccountId != nil {
		t.Fatalf("empty service_account_id must not be sent, got %q", *payload.ServiceAccountId)
	}
}

func TestSetServiceAttributeReadsServiceAccountID(t *testing.T) {
	d := schema.TestResourceDataRaw(t, serviceSchema(), map[string]interface{}{})
	sa := "sa-123"
	service := &koyeb.Service{ServiceAccountId: &sa}
	if err := setServiceAttribute(d, service); err != nil {
		t.Fatalf("setServiceAttribute: %v", err)
	}
	if got := d.Get("service_account_id").(string); got != "sa-123" {
		t.Fatalf("service_account_id = %q, want sa-123", got)
	}
}
