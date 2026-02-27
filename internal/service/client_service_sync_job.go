package service

import (
	"context"
	"sync"
	"time"
)

type clientSyncJob struct {
	syncService ClientSyncService

	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewClientSyncJob creates a clientSyncJob that calls syncService.FullSync on a
// ticker. The job is idle until Start is called.
func NewClientSyncJob(syncService ClientSyncService) ClientSyncJob {
	return &clientSyncJob{syncService: syncService}
}

// Start implements ClientSyncJob. It stops any previously running job, then
// launches a background goroutine that calls FullSync every interval. If interval
// is zero or negative it defaults to 5 minutes. The goroutine exits when ctx is
// cancelled or Stop is called.
func (j *clientSyncJob) Start(ctx context.Context, userID int64, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	j.Stop()

	j.mu.Lock()
	jobCtx, cancel := context.WithCancel(ctx)
	j.cancel = cancel
	j.wg.Add(1)
	j.mu.Unlock()

	go func() {
		defer j.wg.Done()
		t := time.NewTicker(interval)
		defer t.Stop()

		for {
			select {
			case <-jobCtx.Done():
				return
			case <-t.C:
				_ = j.syncService.FullSync(jobCtx, userID)
			}
		}
	}()
}

// Stop implements ClientSyncJob. It cancels the background goroutine's context and
// blocks until the goroutine has fully exited. Safe to call when the job is not
// running (no-op in that case).
func (j *clientSyncJob) Stop() {
	j.mu.Lock()
	cancel := j.cancel
	j.cancel = nil
	j.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	j.wg.Wait()
}
