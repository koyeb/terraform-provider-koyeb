package koyeb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// resolverTestServer serves one app (my-app) and one service per given
// definition; counts let tests assert which listings were walked.
func resolverTestServer(t *testing.T, services string) (*httptest.Server, *int32) {
	t.Helper()
	var lists int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/apps":
			lists++
			_, _ = w.Write([]byte(`{"apps":[{"id":"123e4567-e89b-42d3-a456-426614174000","name":"my-app"}],"count":1}`))
		case "/v1/services":
			lists++
			_, _ = w.Write([]byte(`{"services":[` + services + `],"count":1}`))
		default:
			http.NotFound(w, r)
		}
	}))
	return srv, &lists
}

func resolverTestClient(t *testing.T, srv *httptest.Server) *koyeb.APIClient {
	t.Helper()
	cfg := koyeb.NewConfiguration()
	cfg.Servers[0].URL = srv.URL
	return koyeb.NewAPIClient(cfg)
}

func TestAPIResolverShortCircuitsUUIDRefs(t *testing.T) {
	srv, lists := resolverTestServer(t, "")
	defer srv.Close()

	id, err := newIDResolver(resolverTestClient(t, srv)).App(
		context.Background(), "123e4567-e89b-42d3-a456-426614174001")

	if err != nil || id != "123e4567-e89b-42d3-a456-426614174001" {
		t.Fatalf("expected the UUID ref to short-circuit, got %q, %v", id, err)
	}
	if *lists != 0 {
		t.Errorf("expected no listing for a UUID ref, got %d calls", *lists)
	}
}

func TestAPIResolverAppMatchesNameAndShortID(t *testing.T) {
	srv, _ := resolverTestServer(t, "")
	defer srv.Close()

	resolver := newIDResolver(resolverTestClient(t, srv))
	ctx := context.Background()

	if id, err := resolver.App(ctx, "my-app"); err != nil || id != "123e4567-e89b-42d3-a456-426614174000" {
		t.Errorf("expected the app name to resolve, got %q, %v", id, err)
	}
	// Short IDs are >= 8 character prefixes of the dashless UUID.
	if id, err := resolver.App(ctx, "123e4567"); err != nil || id != "123e4567-e89b-42d3-a456-426614174000" {
		t.Errorf("expected the app short ID to resolve, got %q, %v", id, err)
	}
	_, err := resolver.App(ctx, "no-such-app")
	if err == nil || !strings.Contains(err.Error(), "could not resolve application") {
		t.Errorf("expected an unresolved-ref error, got %v", err)
	}
}

const resolverTestService = `{"id":"123e4567-e89b-42d3-a456-426614174002",` +
	`"name":"my-service","app_id":"123e4567-e89b-42d3-a456-426614174000"}`

const resolverTestDatabase = `{"id":"123e4567-e89b-42d3-a456-426614174003",` +
	`"name":"my-db","app_id":"123e4567-e89b-42d3-a456-426614174000","type":"DATABASE"}`

func TestAPIResolverServiceMatchesSlugAndShortID(t *testing.T) {
	srv, _ := resolverTestServer(t, resolverTestService)
	defer srv.Close()

	resolver := newIDResolver(resolverTestClient(t, srv))
	ctx := context.Background()

	if id, err := resolver.Service(ctx, "my-app/my-service"); err != nil || id != "123e4567-e89b-42d3-a456-426614174002" {
		t.Errorf("expected the app/service slug to resolve, got %q, %v", id, err)
	}
	if id, err := resolver.Service(ctx, "123e4567"); err != nil || id != "123e4567-e89b-42d3-a456-426614174002" {
		t.Errorf("expected the service short ID to resolve, got %q, %v", id, err)
	}
}

func TestAPIResolverDatabaseMatchesCompositeKeys(t *testing.T) {
	srv, _ := resolverTestServer(t, resolverTestDatabase)
	defer srv.Close()

	resolver := newIDResolver(resolverTestClient(t, srv))
	ctx := context.Background()

	for _, ref := range []string{"my-app/my-db", "my-app/123e4567"} {
		if id, err := resolver.Database(ctx, ref); err != nil || id != "123e4567-e89b-42d3-a456-426614174003" {
			t.Errorf("expected the database key %q to resolve, got %q, %v", ref, id, err)
		}
	}
	if id, err := resolver.Database(ctx, "my-app/no-such-db"); err == nil || id != "" {
		t.Errorf("expected an unresolved database ref, got %q, %v", id, err)
	}
}

func TestStaticIDResolverServesRegisteredRefs(t *testing.T) {
	resolver := staticIDResolver{
		"app": {"my-app": "123e4567-e89b-42d3-a456-426614174000"},
	}

	if id, err := resolver.App(context.Background(), "my-app"); err != nil || id != "123e4567-e89b-42d3-a456-426614174000" {
		t.Errorf("expected the registered ref to resolve, got %q, %v", id, err)
	}
	if id, err := resolver.App(context.Background(), "123e4567-e89b-42d3-a456-426614174000"); err != nil {
		t.Errorf("expected a UUID ref to short-circuit, got %q, %v", id, err)
	}
	if _, err := resolver.Service(context.Background(), "my-app/my-service"); err == nil {
		t.Error("expected an unregistered kind/ref to fail")
	}
}
