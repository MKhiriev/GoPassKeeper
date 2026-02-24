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

func NewClientSyncJob(syncService ClientSyncService) ClientSyncJob {
	return &clientSyncJob{syncService: syncService}
}

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
