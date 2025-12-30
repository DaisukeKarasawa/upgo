package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Scheduler struct {
	interval time.Duration
	enabled  bool
	task     func(ctx context.Context) error
	logger   *zap.Logger
	stopCh   chan struct{}
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

	s.logger.Info("スケジューラーを開始しました", zap.Duration("interval", s.interval))

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 初回実行
	go func() {
		if err := s.task(ctx); err != nil {
			s.logger.Error("スケジュールタスクの実行に失敗しました", zap.Error(err))
		}
	}()

	for {
		select {
		case <-ticker.C:
			if err := s.task(ctx); err != nil {
				s.logger.Error("スケジュールタスクの実行に失敗しました", zap.Error(err))
			}
		case <-s.stopCh:
			s.logger.Info("スケジューラーを停止しました")
			return
		case <-ctx.Done():
			s.logger.Info("スケジューラーを停止しました（コンテキスト終了）")
			return
		}
	}
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}
