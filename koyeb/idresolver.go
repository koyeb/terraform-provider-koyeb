package koyeb

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// idResolver turns names and short IDs into Koyeb UUIDs by listing the
// matching collection, mirroring the CLI idmapper semantics it replaces:
// UUIDs short-circuit, otherwise one paged listing matches the dashless
// UUID prefix (>= 8 chars) or the human-readable keys of the kind.
type idResolver interface {
	App(ctx context.Context, ref string) (string, error)
	Service(ctx context.Context, ref string) (string, error)
	Database(ctx context.Context, ref string) (string, error)
	Secret(ctx context.Context, ref string) (string, error)
	Volume(ctx context.Context, ref string) (string, error)
	Snapshot(ctx context.Context, ref string) (string, error)
	Domain(ctx context.Context, ref string) (string, error)
}

// newIDResolver is the resolver seam; production resolves through the API.
var newIDResolver = func(client *koyeb.APIClient) idResolver {
	return &apiIDResolver{client: client}
}

var rxUUIDv4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

const resolverPageSize = 100

// listingPage reads one page of a collection. It reports whether more
// pages follow and whether the page held the looked-up ref.
type listingPage func(offset int64) (more, matched bool, err error)

type apiIDResolver struct {
	client *koyeb.APIClient
}

func (r *apiIDResolver) App(ctx context.Context, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}

	found := ""
	err := resolveFromListing(ctx, "application", ref, []string{
		"the full UUID", "the short ID (8 characters)", "the name",
	}, func(offset int64) (bool, bool, error) {
		res, _, err := r.client.AppsApi.ListApps(ctx).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return false, false, listingError("applications", err)
		}
		for _, app := range res.GetApps() {
			if found == "" && refMatches(ref, app.GetId(), app.GetName()) {
				found = app.GetId()
			}
		}
		return offset+resolverPageSize < res.GetCount(), found != "", nil
	})
	return found, err
}

func (r *apiIDResolver) Service(ctx context.Context, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}

	found := ""
	appNames, err := r.appNames(ctx)
	if err != nil {
		return "", err
	}
	err = resolveFromListing(ctx, "service", ref, []string{
		"the full UUID", "the short ID (8 characters)", "the application name and service name separated by a slash",
	}, func(offset int64) (bool, bool, error) {
		res, _, err := r.client.ServicesApi.ListServices(ctx).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return false, false, listingError("services", err)
		}
		for _, service := range res.GetServices() {
			slug := appNames[service.GetAppId()] + "/" + service.GetName()
			if found == "" && refMatches(ref, service.GetId(), slug) {
				found = service.GetId()
			}
		}
		return offset+resolverPageSize < res.GetCount(), found != "", nil
	})
	return found, err
}

func (r *apiIDResolver) Database(ctx context.Context, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}

	found := ""
	appNames, err := r.appNames(ctx)
	if err != nil {
		return "", err
	}
	err = resolveFromListing(ctx, "database", ref, []string{
		"the full UUID", "the short ID (8 characters)", "the application name and database name separated by a slash",
	}, func(offset int64) (bool, bool, error) {
		res, _, err := r.client.ServicesApi.ListServices(ctx).
			Types([]string{"DATABASE"}).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return false, false, listingError("databases", err)
		}
		for _, service := range res.GetServices() {
			if found == "" && refMatchesDatabase(ref, service, appNames[service.GetAppId()]) {
				found = service.GetId()
			}
		}
		return offset+resolverPageSize < res.GetCount(), found != "", nil
	})
	return found, err
}

func (r *apiIDResolver) Secret(ctx context.Context, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}

	found := ""
	err := resolveFromListing(ctx, "secret", ref, []string{
		"the full UUID", "the short ID (8 characters)", "the name",
	}, func(offset int64) (bool, bool, error) {
		res, _, err := r.client.SecretsApi.ListSecrets(ctx).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return false, false, listingError("secrets", err)
		}
		for _, secret := range res.GetSecrets() {
			if found == "" && refMatches(ref, secret.GetId(), secret.GetName()) {
				found = secret.GetId()
			}
		}
		return offset+resolverPageSize < res.GetCount(), found != "", nil
	})
	return found, err
}

func (r *apiIDResolver) Volume(ctx context.Context, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}

	found := ""
	err := resolveFromListing(ctx, "volume", ref, []string{
		"the full UUID", "the short ID (8 characters)", "the name",
	}, func(offset int64) (bool, bool, error) {
		res, _, err := r.client.PersistentVolumesApi.ListPersistentVolumes(ctx).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return false, false, listingError("volumes", err)
		}
		for _, volume := range res.GetVolumes() {
			if found == "" && refMatches(ref, volume.GetId(), volume.GetName()) {
				found = volume.GetId()
			}
		}
		// Volume listings carry has_next, not a count: a short page can
		// still have a next one when the server caps the page size.
		return res.GetHasNext(), found != "", nil
	})
	return found, err
}

func (r *apiIDResolver) Snapshot(ctx context.Context, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}

	found := ""
	err := resolveFromListing(ctx, "snapshot", ref, []string{
		"the full UUID", "the short ID (8 characters)", "the name",
	}, func(offset int64) (bool, bool, error) {
		res, _, err := r.client.SnapshotsApi.ListSnapshots(ctx).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return false, false, listingError("snapshots", err)
		}
		for _, snapshot := range res.GetSnapshots() {
			if found == "" && refMatches(ref, snapshot.GetId(), snapshot.GetName()) {
				found = snapshot.GetId()
			}
		}
		// Snapshot listings carry has_next, not a count: a short page can
		// still have a next one when the server caps the page size.
		return res.GetHasNext(), found != "", nil
	})
	return found, err
}

func (r *apiIDResolver) Domain(ctx context.Context, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}

	found := ""
	err := resolveFromListing(ctx, "domain", ref, []string{
		"the full UUID", "the short ID (8 characters)", "the name",
	}, func(offset int64) (bool, bool, error) {
		res, _, err := r.client.DomainsApi.ListDomains(ctx).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return false, false, listingError("domains", err)
		}
		for _, domain := range res.GetDomains() {
			if found == "" && refMatches(ref, domain.GetId(), domain.GetName()) {
				found = domain.GetId()
			}
		}
		return offset+resolverPageSize < res.GetCount(), found != "", nil
	})
	return found, err
}

// appNames lists the organization's apps once and indexes them by ID, for
// the composite service and database keys.
func (r *apiIDResolver) appNames(ctx context.Context) (map[string]string, error) {
	names := map[string]string{}
	offset := int64(0)
	for {
		res, _, err := r.client.AppsApi.ListApps(ctx).
			Limit(strconv.FormatInt(resolverPageSize, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return nil, listingError("applications", err)
		}
		for _, app := range res.GetApps() {
			names[app.GetId()] = app.GetName()
		}
		offset += resolverPageSize
		if offset >= res.GetCount() {
			return names, nil
		}
	}
}

// resolveFromListing short-circuits UUIDs and otherwise walks the kind's
// listing page by page until a page reports the match or the pages end.
func resolveFromListing(ctx context.Context, kind, ref string, forms []string, page listingPage) error {
	offset := int64(0)
	for {
		more, matched, err := page(offset)
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
		if !more {
			break
		}
		offset += resolverPageSize
	}
	return fmt.Errorf("could not resolve %s %q: expected %s", kind, ref, strings.Join(forms, ", "))
}

// refMatches reports whether ref is the resource's name, its app/name
// slug, or a >= 8 character prefix of its dashless UUID.
func refMatches(ref, id, name string) bool {
	return ref == name || isShortID(ref, id)
}

// findServiceByName lists the app's services and returns the id of the one
// with the given name, for adopting an orphan left by a crashed apply.
func findServiceByName(ctx context.Context, client *koyeb.APIClient, appID, name string) (string, bool, error) {
	req := client.ServicesApi.ListServices(ctx).Name(name)
	if appID != "" {
		req = req.AppId(appID)
	}
	res, _, err := req.Execute()
	if err != nil {
		return "", false, listingError("services", err)
	}
	for _, s := range res.GetServices() {
		if s.GetName() == name {
			return s.GetId(), true, nil
		}
	}
	return "", false, nil
}

// findPoolByName lists pools and returns the id of the one with the given
// name, for adopting an orphan left by a crashed apply.
func findPoolByName(ctx context.Context, client *koyeb.APIClient, name string) (string, bool, error) {
	res, _, err := client.ServicePoolsApi.ListServicePools(ctx).Name(name).Execute()
	if err != nil {
		return "", false, listingError("service pools", err)
	}
	for _, p := range res.GetServicePools() {
		if p.GetName() == name {
			return p.GetId(), true, nil
		}
	}
	return "", false, nil
}

// refMatchesDatabase mirrors the CLI's nine acceptable database keys:
// app name, app UUID or app short ID, each combined with the database
// UUID, short ID or name.
func refMatchesDatabase(ref string, service koyeb.ServiceListItem, appName string) bool {
	id := service.GetId()
	if isShortID(ref, id) {
		return true
	}
	dashless := strings.ReplaceAll(id, "-", "")
	sid := dashless
	if len(sid) > 8 {
		sid = sid[:8]
	}
	appID := service.GetAppId()
	appShort := appID
	if len(appShort) > 8 {
		appShort = appShort[:8]
	}
	for _, key := range []string{
		appName + "/" + id, appID + "/" + id, appShort + "/" + id,
		appName + "/" + sid, appID + "/" + sid, appShort + "/" + sid,
		appName + "/" + service.GetName(), appID + "/" + service.GetName(), appShort + "/" + service.GetName(),
	} {
		if ref == key {
			return true
		}
	}
	return false
}

func isShortID(ref, id string) bool {
	return len(ref) >= 8 && strings.HasPrefix(strings.ReplaceAll(id, "-", ""), ref)
}

func listingError(kind string, err error) error {
	return fmt.Errorf("error listing %s to resolve the identifier: %s", kind, err)
}

// staticIDResolver is the in-memory adapter for tests: pre-registered
// refs per kind, UUIDs short-circuit like the API adapter.
type staticIDResolver map[string]map[string]string

func (s staticIDResolver) resolve(kind, ref string) (string, error) {
	if rxUUIDv4.MatchString(ref) {
		return ref, nil
	}
	if id, ok := s[kind][ref]; ok {
		return id, nil
	}
	return "", fmt.Errorf("could not resolve %s %q", kind, ref)
}

func (s staticIDResolver) App(_ context.Context, ref string) (string, error) {
	return s.resolve("app", ref)
}

func (s staticIDResolver) Service(_ context.Context, ref string) (string, error) {
	return s.resolve("service", ref)
}

func (s staticIDResolver) Database(_ context.Context, ref string) (string, error) {
	return s.resolve("database", ref)
}

func (s staticIDResolver) Secret(_ context.Context, ref string) (string, error) {
	return s.resolve("secret", ref)
}

func (s staticIDResolver) Volume(_ context.Context, ref string) (string, error) {
	return s.resolve("volume", ref)
}

func (s staticIDResolver) Snapshot(_ context.Context, ref string) (string, error) {
	return s.resolve("snapshot", ref)
}

func (s staticIDResolver) Domain(_ context.Context, ref string) (string, error) {
	return s.resolve("domain", ref)
}
