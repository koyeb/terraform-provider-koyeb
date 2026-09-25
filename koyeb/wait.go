package koyeb

import (
	"context"
	"errors"
	"fmt"
	_nethttp "net/http"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"golang.org/x/exp/slices"
)

// Poll interval between readiness checks; a variable so tests can shorten
// it. Mutating tests must not use t.Parallel().
var waitRetryInterval = 5 * time.Second

// errGone marks a 404: the resource no longer exists.
var errGone = errors.New("resource is gone")

// errTransient marks a network failure or 5xx worth retrying until the
// deadline when the wait allows it; the underlying error is wrapped.
var errTransient = errors.New("transient API failure")

// statusWait describes one readiness or destroy wait. The loop hides in
// waitForStatus; poll adapters translate the generated replies.
type statusWait struct {
	name      string
	targets   []string
	terminals []string
	timeout   time.Duration
	interval  time.Duration
	// notFoundGone counts a 404 as success (destroy checks).
	notFoundGone bool
	// retryTransient keeps polling transient failures until the deadline.
	retryTransient bool
}

// waitForStatus polls until the resource reaches a target status, fails
// fast on terminal statuses, and surfaces cancellations; unknown statuses
// keep polling until the timeout names them.
func waitForStatus(ctx context.Context, w statusWait, poll func() (string, error)) error {
	lastStatus := "unknown"
	deadline := time.Now().Add(w.timeout)

	for time.Now().Before(deadline) {
		status, err := poll()
		switch {
		case err == nil:
			lastStatus = status
			if slices.Contains(w.targets, status) {
				return nil
			}
			if slices.Contains(w.terminals, status) {
				return fmt.Errorf("%s reached terminal status %s", w.name, status)
			}
		case errors.Is(err, errGone) && w.notFoundGone:
			return nil
		case w.retryTransient && errors.Is(err, errTransient) && time.Now().Before(deadline):
			// Retry the blip after the interval.
		default:
			return err
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for %s cancelled: %w", w.name, ctx.Err())
		case <-time.After(w.interval):
		}
	}

	return fmt.Errorf("wait for %s timed out after %s (last status %s)", w.name, w.timeout, lastStatus)
}

// goneWait describes a destroy check: a 404 or a target status ends it.
func goneWait(name string, targets []string, timeout time.Duration) statusWait {
	return statusWait{
		name:         name,
		targets:      targets,
		timeout:      timeout,
		interval:     waitRetryInterval,
		notFoundGone: true,
	}
}

// pollReply translates one generated-client reply into a status or a
// classified error: 404 becomes errGone, network failures and 5xx become
// errTransient, everything else aborts the wait.
func pollReply[T any](res T, resp *_nethttp.Response, err error, status func(T) string) (string, error) {
	if err != nil {
		if resp != nil && resp.StatusCode == _nethttp.StatusNotFound {
			return "", errGone
		}
		if resp == nil || resp.StatusCode >= 500 {
			return "", fmt.Errorf("%w: %w", errTransient, err)
		}
		return "", err
	}
	return status(res), nil
}

func serviceStatusPoller(ctx context.Context, client *koyeb.APIClient, id string) func() (string, error) {
	return func() (string, error) {
		res, resp, err := client.ServicesApi.GetService(ctx, id).Execute()
		return pollReply(res, resp, err, func(r *koyeb.GetServiceReply) string {
			service := r.GetService()
			return string(service.GetStatus())
		})
	}
}

func servicePoolStatusPoller(ctx context.Context, client *koyeb.APIClient, id string) func() (string, error) {
	return func() (string, error) {
		res, resp, err := client.ServicePoolsApi.GetServicePool(ctx, id).Execute()
		return pollReply(res, resp, err, func(r *koyeb.GetServicePoolReply) string {
			pool := r.GetServicePool()
			return string(pool.GetStatus())
		})
	}
}

func poolClaimStatusPoller(ctx context.Context, client *koyeb.APIClient, claimID string) func() (string, error) {
	return func() (string, error) {
		res, resp, err := client.PoolClaimsApi.GetClaim(ctx, claimID).Execute()
		status, pollErr := pollReply(res, resp, err, func(r *koyeb.GetPoolClaimReply) string {
			claim := r.GetClaim()
			return string(claim.GetStatus())
		})
		if pollErr != nil && errors.Is(pollErr, errGone) {
			return "", fmt.Errorf("claim %s disappeared while waiting for fulfillment: %w", claimID, pollErr)
		}
		return status, pollErr
	}
}

func appStatusPoller(ctx context.Context, client *koyeb.APIClient, id string) func() (string, error) {
	return func() (string, error) {
		res, resp, err := client.AppsApi.GetApp(ctx, id).Execute()
		return pollReply(res, resp, err, func(r *koyeb.GetAppReply) string {
			app := r.GetApp()
			return string(app.GetStatus())
		})
	}
}

func domainStatusPoller(ctx context.Context, client *koyeb.APIClient, id string) func() (string, error) {
	return func() (string, error) {
		res, resp, err := client.DomainsApi.GetDomain(ctx, id).Execute()
		return pollReply(res, resp, err, func(r *koyeb.GetDomainReply) string {
			domain := r.GetDomain()
			return string(domain.GetStatus())
		})
	}
}

func volumeStatusPoller(ctx context.Context, client *koyeb.APIClient, id string) func() (string, error) {
	return func() (string, error) {
		res, resp, err := client.PersistentVolumesApi.GetPersistentVolume(ctx, id).Execute()
		return pollReply(res, resp, err, func(r *koyeb.GetPersistentVolumeReply) string {
			volume := r.GetVolume()
			return string(volume.GetStatus())
		})
	}
}

func snapshotStatusPoller(ctx context.Context, client *koyeb.APIClient, id string) func() (string, error) {
	return func() (string, error) {
		res, resp, err := client.SnapshotsApi.GetSnapshot(ctx, id).Execute()
		return pollReply(res, resp, err, func(r *koyeb.GetSnapshotReply) string {
			snapshot := r.GetSnapshot()
			return string(snapshot.GetStatus())
		})
	}
}
