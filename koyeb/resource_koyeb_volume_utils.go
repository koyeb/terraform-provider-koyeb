package koyeb

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func isStillAttachedError(err error) bool {
	var apiErr *koyeb.GenericOpenAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return strings.Contains(string(apiErr.Body()), "still attached")
}

// deleteVolumeWhenDetached deletes a volume, tolerating the window where a
// service that mounted it is still being torn down: the API rejects the
// delete with "still attached" until the detach completes.
func deleteVolumeWhenDetached(client *koyeb.APIClient, id string, interval time.Duration) error {
	const attempts = 30

	for i := 0; i < attempts; i++ {
		_, resp, err := client.PersistentVolumesApi.DeletePersistentVolume(context.Background(), id).Execute()
		if err == nil {
			return nil
		}
		if resp == nil || resp.StatusCode != 400 || !isStillAttachedError(err) {
			return err
		}
		time.Sleep(interval)
	}

	return errors.New("volume is still attached after waiting for the service to detach")
}
