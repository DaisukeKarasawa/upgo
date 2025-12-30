package scheduler

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Scheduler struct {
	interval time.Duration
	enabled  bool
	task     func(ctx context.Context) error
	logger   *zap.Logger
	stopCh   chan struct{}
	// wg tracks running goroutines for graceful shutdown
	wg sync.WaitGroup
	// mu protects the running state
	mu sync.Mutex
	// running indicates if the scheduler is currently running
	running bool
}

func NewScheduler(interval string, enabled bool, task func(ctx context.Context) error, logger *zap.Logger) (*Scheduler, error) {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		interval: duration,
		enabled:  enabled,
		task:     task,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) {
	if !s.enabled {
		s.logger.Info("スケジューラーは無効になっています")
		return
	}

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		s.logger.Warn("スケジューラーは既に実行中です")
		return
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("スケジューラーを開始しました", zap.Duration("interval", s.interval))

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Execute first time asynchronously
	s.wg.Add(1)
	go s.runTask(ctx)

	for {
		select {
		case <-ticker.C:
			// Execute task asynchronously to prevent blocking the ticker
			s.wg.Add(1)
			go s.runTask(ctx)
		case <-s.stopCh:
			s.logger.Info("スケジューラーを停止しました")
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
			// Wait for all running tasks to complete
			s.wg.Wait()
			return
		case <-ctx.Done():
			s.logger.Info("スケジューラーを停止しました（コンテキスト終了）")
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
			// Wait for all running tasks to complete
			s.wg.Wait()
			return
		}
	}
}

// runTask executes the scheduled task in a goroutine and tracks it with WaitGroup.
func (s *Scheduler) runTask(ctx context.Context) {
	defer s.wg.Done()
	if err := s.task(ctx); err != nil {
		if ctx.Err() == context.Canceled {
			s.logger.Info("スケジュールタスクがキャンセルされました")
		} else {
			s.logger.Error("スケジュールタスクの実行に失敗しました", zap.Error(err))
		}
	}
}

// Stop stops the scheduler and waits for all running tasks to complete.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	close(s.stopCh)
	// Wait for all running tasks to complete
	s.wg.Wait()
}
