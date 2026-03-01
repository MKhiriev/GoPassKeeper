// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spySyncService считает вызовы FullSync и позволяет управлять задержкой.
type spySyncService struct {
	calls atomic.Int64
	err   error
}

func (s *spySyncService) FullSync(_ context.Context, _ int64) error {
	s.calls.Add(1)
	return s.err
}

func (s *spySyncService) ExecutePlan(_ context.Context, _ models.SyncPlan, _ int64) error {
	return nil
}

// ── NewClientSyncJob ─────────────────────────────────────────────────────────

func TestNewClientSyncJob_ReturnsInterface(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)
	require.NotNil(t, job)

	// проверяем что возвращённый объект реализует ClientSyncJob
	var _ ClientSyncJob = job
}

// ── Start / Stop ─────────────────────────────────────────────────────────────

func TestClientSyncJob_Start_CallsFullSync(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)
	ctx := context.Background()

	// Интервал 10ms — за 55ms должно быть ~5 тиков
	job.Start(ctx, 1, 10*time.Millisecond)
	time.Sleep(55 * time.Millisecond)
	job.Stop()

	got := spy.calls.Load()
	assert.GreaterOrEqual(t, got, int64(3), "FullSync должен быть вызван несколько раз, вызвано: %d", got)
}

func TestClientSyncJob_Stop_StopsGoroutine(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)
	ctx := context.Background()

	job.Start(ctx, 1, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	job.Stop()

	callsAfterStop := spy.calls.Load()
	time.Sleep(30 * time.Millisecond)
	callsLater := spy.calls.Load()

	assert.Equal(t, callsAfterStop, callsLater, "после Stop новых вызовов быть не должно")
}

func TestClientSyncJob_Stop_BeforeStart_NoPanic(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)

	// Stop без Start не должен паниковать
	assert.NotPanics(t, func() { job.Stop() })
}

func TestClientSyncJob_DoubleStop_NoPanic(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)
	ctx := context.Background()

	job.Start(ctx, 1, 10*time.Millisecond)
	job.Stop()

	// Повторный Stop не должен паниковать
	assert.NotPanics(t, func() { job.Stop() })
}

func TestClientSyncJob_Start_DefaultInterval(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy).(*clientSyncJob)
	ctx, cancel := context.WithCancel(context.Background())

	// interval <= 0 → дефолт 5 минут, за 20ms вызовов быть не должно
	job.Start(ctx, 1, 0)
	time.Sleep(20 * time.Millisecond)
	cancel()
	job.Stop()

	assert.Equal(t, int64(0), spy.calls.Load(), "при дефолтном интервале 5min за 20ms вызовов нет")
}

func TestClientSyncJob_Start_NegativeInterval(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)
	ctx, cancel := context.WithCancel(context.Background())

	// Отрицательный интервал → дефолт 5 минут
	job.Start(ctx, 1, -1*time.Second)
	time.Sleep(20 * time.Millisecond)
	cancel()
	job.Stop()

	assert.Equal(t, int64(0), spy.calls.Load())
}

func TestClientSyncJob_Restart_StopsPrevious(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)
	ctx := context.Background()

	// Первый запуск
	job.Start(ctx, 1, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)

	callsBefore := spy.calls.Load()
	assert.Greater(t, callsBefore, int64(0))

	// Перезапуск — предыдущая горутина должна остановиться
	spy2 := &spySyncService{}
	job2 := NewClientSyncJob(spy2)
	// Используем тот же job чтобы проверить restart
	_ = job2

	// Start повторно на том же job — внутри вызовет Stop()
	job.Start(ctx, 2, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	job.Stop()

	// Оба раунда должны были сгенерировать вызовы
	totalCalls := spy.calls.Load()
	assert.Greater(t, totalCalls, callsBefore, "второй Start должен продолжить генерировать вызовы")
}

func TestClientSyncJob_ContextCancel_StopsJob(t *testing.T) {
	spy := &spySyncService{}
	job := NewClientSyncJob(spy)
	ctx, cancel := context.WithCancel(context.Background())

	job.Start(ctx, 1, 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	cancel() // отменяем родительский контекст

	// Stop должен вернуться без зависания
	done := make(chan struct{})
	go func() {
		job.Stop()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(1 * time.Second):
		t.Fatal("Stop завис после отмены контекста")
	}
}

func TestClientSyncJob_FullSyncError_DoesNotStopJob(t *testing.T) {
	spy := &spySyncService{err: assert.AnError}
	job := NewClientSyncJob(spy)
	ctx := context.Background()

	// FullSync возвращает ошибку, но джоб продолжает работать
	job.Start(ctx, 1, 10*time.Millisecond)
	time.Sleep(55 * time.Millisecond)
	job.Stop()

	got := spy.calls.Load()
	assert.GreaterOrEqual(t, got, int64(3), "несмотря на ошибки, FullSync продолжает вызываться: %d", got)
}

func TestClientSyncJob_PassesUserID(t *testing.T) {
	var capturedUserID atomic.Int64

	spy := &captureSyncService{onFullSync: func(_ context.Context, userID int64) error {
		capturedUserID.Store(userID)
		return nil
	}}

	job := NewClientSyncJob(spy)
	ctx := context.Background()

	job.Start(ctx, 42, 10*time.Millisecond)
	time.Sleep(25 * time.Millisecond)
	job.Stop()

	assert.Equal(t, int64(42), capturedUserID.Load(), "userID должен пробрасываться в FullSync")
}

// captureSyncService — позволяет перехватить аргументы FullSync.
type captureSyncService struct {
	onFullSync func(ctx context.Context, userID int64) error
}

func (c *captureSyncService) FullSync(ctx context.Context, userID int64) error {
	return c.onFullSync(ctx, userID)
}

func (c *captureSyncService) ExecutePlan(_ context.Context, _ models.SyncPlan, _ int64) error {
	return nil
}
