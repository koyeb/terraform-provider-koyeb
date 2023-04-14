package koyeb

import (
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

func waitForResourceStatus[T any](fn func() (T, *_nethttp.Response, error), resourceName string, targetStatus []string, timeout time.Duration, throwErrorIfNotFound bool) error {
	var status string
	now := time.Now()
	retryInterval := 5 * time.Second
	timeoutAt := time.Minute * timeout

	for time.Since(now) < timeoutAt {
		res, resp, err := fn()
		if err != nil {
			if resp.StatusCode == 404 && !throwErrorIfNotFound {
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
		default:
			return errors.New("unknown resource type")
		}

		if slices.Contains(targetStatus, status) {
			return nil
		}
		time.Sleep(retryInterval)
	}

	return errors.New("resource failed to reach target status after timeout")
}
