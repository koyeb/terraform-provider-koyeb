package koyeb

import (
	"context"
	"errors"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// deleteVolumeWhenDetached deletes a volume, tolerating the window where a
// service that mounted it is still being torn down: the API rejects the
// delete with a bare 400 until the detach completes, so the status — not the
// body text — is the retry signal.
func deleteVolumeWhenDetached(ctx context.Context, client *koyeb.APIClient, id string, interval time.Duration) error {
	const attempts = 30

	for i := 0; i < attempts; i++ {
		_, resp, err := client.PersistentVolumesApi.DeletePersistentVolume(ctx, id).Execute()
		if err == nil {
			return nil
		}
		// A retried destroy after a crash finds the volume already gone.
		if isGone(resp) {
			return nil
		}
		if resp == nil || resp.StatusCode != 400 {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}

	return errors.New("volume is still attached after waiting for the service to detach")
}
