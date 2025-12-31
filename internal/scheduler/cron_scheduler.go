package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// CronScheduler supports cron format scheduling (e.g., "0 0 * * *" for daily at 00:00)
type CronScheduler struct {
	spec    string
	enabled bool
	task    func(ctx context.Context) error
	logger  *zap.Logger
	cron    *cron.Cron
	// wg tracks running goroutines for graceful shutdown
	wg sync.WaitGroup
	// mu protects the running state
	mu sync.Mutex
	// running indicates if the scheduler is currently running
	running bool
	// entryID stores the cron entry ID
	entryID cron.EntryID
}

// NewCronScheduler creates a new cron scheduler
// spec: cron expression (e.g., "0 0 * * *" for daily at 00:00)
func NewCronScheduler(spec string, enabled bool, task func(ctx context.Context) error, logger *zap.Logger) (*CronScheduler, error) {
	// Validate cron spec
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(spec); err != nil {
		return nil, err
	}

	// Use standard 5-field cron format (minute hour day month weekday)
	c := cron.New()
	if !enabled {
		c.Stop()
	}

	return &CronScheduler{
		spec:    spec,
		enabled: enabled,
		task:    task,
		logger:  logger,
		cron:    c,
	}, nil
}

// Start starts the cron scheduler
func (s *CronScheduler) Start(ctx context.Context) {
	if !s.enabled {
		s.logger.Info("Cronスケジューラーは無効になっています")
		return
	}

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		s.logger.Warn("Cronスケジューラーは既に実行中です")
		return
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Cronスケジューラーを開始しました", zap.String("spec", s.spec))

	// Add cron job with standard 5-field format parser
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	entryID, err := s.cron.AddFunc(s.spec, func() {
		s.wg.Add(1)
		go s.runTask(ctx)
	})
	if err != nil {
		s.logger.Error("Cronジョブの追加に失敗しました", zap.Error(err))
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		return
	}
	s.entryID = entryID

	// Parse to get next run time for logging
	schedule, err := parser.Parse(s.spec)
	if err == nil {
		now := time.Now()
		nextRun := schedule.Next(now)
		timeUntilNext := nextRun.Sub(now)
		s.logger.Info("次回実行予定時刻", 
			zap.Time("next_run", nextRun),
			zap.Duration("time_until_next", timeUntilNext),
		)
	}

	// Start cron scheduler
	s.cron.Start()

	// Wait for context cancellation or stop signal
	go func() {
		<-ctx.Done()
		s.Stop()
	}()
}

// runTask executes the scheduled task in a goroutine and tracks it with WaitGroup.
func (s *CronScheduler) runTask(ctx context.Context) {
	defer s.wg.Done()
	if err := s.task(ctx); err != nil {
		if ctx.Err() == context.Canceled {
			s.logger.Info("Cronスケジュールタスクがキャンセルされました")
		} else {
			s.logger.Error("Cronスケジュールタスクの実行に失敗しました", zap.Error(err))
		}
	}
}

// Stop stops the cron scheduler and waits for all running tasks to complete.
func (s *CronScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	s.cron.Stop()
	s.logger.Info("Cronスケジューラーを停止しました")
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
	// Wait for all running tasks to complete
	s.wg.Wait()
}

// IsCronSpec checks if a string is a cron spec (simple heuristic)
func IsCronSpec(spec string) bool {
	// Basic check: if it contains spaces and numbers, it might be cron
	// More sophisticated check would parse it
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	_, err := parser.Parse(spec)
	return err == nil
}
