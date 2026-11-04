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

func TestRefMatchesDatabaseToleratesShortIDs(t *testing.T) {
	id, appID, name := "abc", "xy", "my-db"
	service := koyeb.ServiceListItem{Id: &id, AppId: &appID, Name: &name}
	if !refMatchesDatabase("my-app/my-db", service, "my-app") {
		t.Error("expected the app/name key to match")
	}
	// Short or malformed API ids must not panic the provider.
	if refMatchesDatabase("my-app/xyz", service, "my-app") {
		t.Error("an unknown key must not match")
	}
}

func TestFindByNamePropagatesListErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	}))
	defer srv.Close()
	client := resolverTestClient(t, srv)
	if _, _, err := findServiceByName(context.Background(), client, "app-id", "svc"); err == nil {
		t.Error("expected the service listing error to propagate, got nil")
	}
	if _, _, err := findPoolByName(context.Background(), client, "pool"); err == nil {
		t.Error("expected the pool listing error to propagate, got nil")
	}
}

// The resolver must walk listings page by page until the reported count is
// covered and keep the first match when several items match.
func TestAPIResolverServicePaginatesAndKeepsFirstMatch(t *testing.T) {
	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/apps":
			_, _ = w.Write([]byte(`{"apps":[{"id":"app-1","name":"my-app"}],"count":1}`))
		case "/v1/services":
			requests++
			offset := r.URL.Query().Get("offset")
			switch offset {
			case "0":
				_, _ = w.Write([]byte(`{"services":[{"id":"svc-page0","name":"other","app_id":"app-1"}],"count":200,"has_next":true}`))
			default:
				_, _ = w.Write([]byte(`{"services":[{"id":"svc-page1","name":"dup","app_id":"app-1"},{"id":"svc-page1-b","name":"dup","app_id":"app-1"}],"count":200}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	resolver := newIDResolver(resolverTestClient(t, srv))
	id, err := resolver.Service(context.Background(), "my-app/dup")
	if err != nil {
		t.Fatalf("expected the slug on page 2 to resolve, got %v", err)
	}
	if id != "svc-page1" {
		t.Errorf("expected the first match on the page to win, got %q", id)
	}
	if requests != 2 {
		t.Errorf("expected exactly 2 listing pages, got %d requests", requests)
	}
}

// Volume/snapshot listings carry has_next instead of a count: a short page
// with has_next=true must not end the walk early.
func TestAPIResolverVolumeFollowsHasNextOnShortPage(t *testing.T) {
	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/v1/volumes" {
			http.NotFound(w, r)
			return
		}
		requests++
		if r.URL.Query().Get("offset") == "0" {
			// Short page (1 < limit 100) but has_next=true; the walk must
			// resume right after the received item, not skip to offset 100.
			_, _ = w.Write([]byte(`{"volumes":[{"id":"vol-page0","name":"other"}],"has_next":true}`))
			return
		}
		if r.URL.Query().Get("offset") != "1" {
			t.Errorf("expected the next page at offset 1 (items received), got %q", r.URL.Query().Get("offset"))
		}
		_, _ = w.Write([]byte(`{"volumes":[{"id":"vol-page1","name":"my-vol"}],"has_next":false}`))
	}))
	defer srv.Close()

	id, err := newIDResolver(resolverTestClient(t, srv)).Volume(context.Background(), "my-vol")
	if err != nil {
		t.Fatalf("expected the volume on page 2 to resolve, got %v", err)
	}
	if id != "vol-page1" {
		t.Errorf("expected vol-page1, got %q", id)
	}
	if requests != 2 {
		t.Errorf("expected 2 listing requests, got %d", requests)
	}
}

// A server-capped short page must not end the walk early for the
// count-based kinds either: has_next decides, the offset advances by the
// items actually received.
func TestAPIResolverAppFollowsHasNextOnShortPage(t *testing.T) {
	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/v1/apps" {
			http.NotFound(w, r)
			return
		}
		requests++
		if r.URL.Query().Get("offset") == "0" {
			_, _ = w.Write([]byte(`{"apps":[{"id":"app-0","name":"other"}],"count":250,"has_next":true}`))
			return
		}
		if r.URL.Query().Get("offset") != "1" {
			t.Errorf("expected the next page at offset 1, got %q", r.URL.Query().Get("offset"))
		}
		_, _ = w.Write([]byte(`{"apps":[{"id":"app-1","name":"my-app"}],"count":250,"has_next":false}`))
	}))
	defer srv.Close()

	id, err := newIDResolver(resolverTestClient(t, srv)).App(context.Background(), "my-app")
	if err != nil {
		t.Fatalf("expected the app on page 2 to resolve, got %v", err)
	}
	if id != "app-1" {
		t.Errorf("expected app-1, got %q", id)
	}
	if requests != 2 {
		t.Errorf("expected 2 listing requests, got %d", requests)
	}
}
