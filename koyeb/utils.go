package koyeb

import (
	"context"
	"fmt"
	_nethttp "net/http"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"golang.org/x/exp/slices"
)

func toOpt[T any](v T) *T {
	return &v
}

// Poll interval between readiness checks; a variable so tests can shorten
// it. Mutating tests must not use t.Parallel().
var waitRetryInterval = 5 * time.Second

// waitForResourceStatus polls fn until the resource reaches targetStatus,
// failing fast when it lands in any of the optional terminalStatuses;
// unknown statuses keep polling until the timeout names them.
func waitForResourceStatus[T any](ctx context.Context, fn func() (T, *_nethttp.Response, error), resourceName string, targetStatus []string, timeout time.Duration, throwErrorIfNotFound bool, terminalStatuses ...string) error {
	lastStatus := "unknown"
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		res, resp, err := fn()
		if err != nil {
			if resp != nil && resp.StatusCode == 404 && !throwErrorIfNotFound {
				return nil
			}
			return err
		}

		var status string
		switch v := any(res).(type) {
		case *koyeb.GetServiceReply:
			status = string(v.Service.GetStatus())
		case *koyeb.GetDeploymentReply:
			status = string(v.Deployment.GetStatus())
		case *koyeb.GetDomainReply:
			status = string(v.Domain.GetStatus())
		case *koyeb.GetServicePoolReply:
			status = string(v.ServicePool.GetStatus())
		case *koyeb.GetPersistentVolumeReply:
			status = string(v.Volume.GetStatus())
		case *koyeb.GetSnapshotReply:
			status = string(v.Snapshot.GetStatus())
		case *koyeb.GetAppReply:
			status = string(v.App.GetStatus())
		default:
			return fmt.Errorf("unknown resource type for wait on %s", resourceName)
		}

		lastStatus = status
		if slices.Contains(targetStatus, status) {
			return nil
		}
		if slices.Contains(terminalStatuses, status) {
			return fmt.Errorf("%s reached terminal status %s", resourceName, status)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for %s cancelled: %w", resourceName, ctx.Err())
		case <-time.After(waitRetryInterval):
		}
	}

	return fmt.Errorf("wait for %s timed out after %s (last status %s)", resourceName, timeout, lastStatus)
}
