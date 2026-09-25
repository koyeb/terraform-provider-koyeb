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

func toOpt[T any](v T) *T {
	return &v
}

// Poll interval between readiness checks; a variable so tests can shorten it.
var waitRetryInterval = 5 * time.Second

func waitForResourceStatus[T any](ctx context.Context, fn func() (T, *_nethttp.Response, error), resourceName string, targetStatus []string, timeout time.Duration, throwErrorIfNotFound bool) error {
	var status string
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		res, resp, err := fn()
		if err != nil {
			if resp != nil && resp.StatusCode == 404 && !throwErrorIfNotFound {
				return nil
			}
			return err
		}

		switch v := any(res).(type) {
		case *koyeb.GetServiceReply:
			status = fmt.Sprintf("%v", v.Service.GetStatus())
		case *koyeb.GetDeploymentReply:
			status = fmt.Sprintf("%v", v.Deployment.GetStatus())
		case *koyeb.GetDomainReply:
			status = fmt.Sprintf("%v", v.Domain.GetStatus())
		case *koyeb.GetServicePoolReply:
			status = fmt.Sprintf("%v", v.ServicePool.GetStatus())
		case *koyeb.GetPersistentVolumeReply:
			status = fmt.Sprintf("%v", v.Volume.GetStatus())
		case *koyeb.GetSnapshotReply:
			status = fmt.Sprintf("%v", v.Snapshot.GetStatus())
		case *koyeb.GetAppReply:
			status = fmt.Sprintf("%v", v.App.GetStatus())
		default:
			return errors.New("unknown resource type")
		}

		if slices.Contains(targetStatus, status) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for %s cancelled: %w", resourceName, ctx.Err())
		case <-time.After(waitRetryInterval):
		}
	}

	return errors.New("resource failed to reach target status after timeout")
}
