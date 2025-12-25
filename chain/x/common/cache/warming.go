// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cosmossdk.io/log"
)

// WarmupStrategy defines how cache warming should be performed
type WarmupStrategy string

const (
	WarmupSequential WarmupStrategy = "sequential"
	WarmupParallel   WarmupStrategy = "parallel"
	WarmupLazy       WarmupStrategy = "lazy"
)

// Warmer handles cache warming operations
type Warmer struct {
	cache    *Cache
	logger   log.Logger
	strategy WarmupStrategy
	workers  int
}

// NewWarmer creates a new cache warmer
func NewWarmer(cache *Cache, logger log.Logger, strategy WarmupStrategy, workers int) *Warmer {
	if workers <= 0 {
		workers = 4
	}
	return &Warmer{
		cache:    cache,
		logger:   logger,
		strategy: strategy,
		workers:  workers,
	}
}

// WarmupTask represents a single cache warming task
type WarmupTask struct {
	Key    string
	Loader func() (interface{}, error)
	TTL    time.Duration
}

// Warmup warms up the cache with the provided tasks
func (w *Warmer) Warmup(ctx context.Context, tasks []WarmupTask) error {
	w.logger.Info("starting cache warmup",
		"tasks", len(tasks),
		"strategy", w.strategy)

	start := time.Now()

	var err error
	switch w.strategy {
	case WarmupSequential:
		err = w.warmupSequential(ctx, tasks)
	case WarmupParallel:
		err = w.warmupParallel(ctx, tasks)
	case WarmupLazy:
		err = w.warmupLazy(ctx, tasks)
	default:
		return fmt.Errorf("unknown warmup strategy: %s", w.strategy)
	}

	if err != nil {
		return fmt.Errorf("warmup failed: %w", err)
	}

	duration := time.Since(start)
	w.logger.Info("cache warmup completed",
		"duration", duration,
		"tasks", len(tasks))

	return nil
}

// warmupSequential warms up cache sequentially
func (w *Warmer) warmupSequential(ctx context.Context, tasks []WarmupTask) error {
	for i, task := range tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := w.executeTask(task); err != nil {
			w.logger.Error("warmup task failed",
				"task", i,
				"key", task.Key,
				"error", err)
			continue
		}
	}
	return nil
}

// warmupParallel warms up cache in parallel
func (w *Warmer) warmupParallel(ctx context.Context, tasks []WarmupTask) error {
	taskChan := make(chan WarmupTask, len(tasks))
	errChan := make(chan error, w.workers)
	var wg sync.WaitGroup

	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				if err := w.executeTask(task); err != nil {
					w.logger.Error("warmup task failed",
						"worker", workerID,
						"key", task.Key,
						"error", err)
					errChan <- err
				}
			}
		}(i)
	}

	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}

	return nil
}

// warmupLazy schedules lazy warming
func (w *Warmer) warmupLazy(ctx context.Context, tasks []WarmupTask) error {
	go func() {
		time.Sleep(5 * time.Second)
		if err := w.warmupParallel(ctx, tasks); err != nil {
			w.logger.Error("lazy warmup failed", "error", err)
		}
	}()
	return nil
}

// executeTask executes a single warmup task
func (w *Warmer) executeTask(task WarmupTask) error {
	value, err := task.Loader()
	if err != nil {
		return fmt.Errorf("loader failed for key %s: %w", task.Key, err)
	}

	ttl := task.TTL
	if ttl == 0 {
		ttl = w.cache.config.DefaultTTL
	}

	if err := w.cache.Set(task.Key, value, ttl); err != nil {
		return fmt.Errorf("cache set failed for key %s: %w", task.Key, err)
	}

	return nil
}

// PredictiveWarmer predicts and warms frequently accessed keys
type PredictiveWarmer struct {
	cache        *Cache
	logger       log.Logger
	accessLog    map[string]*accessStats
	mu           sync.RWMutex
	warmThreshold int
}

type accessStats struct {
	count      int
	lastAccess time.Time
}

// NewPredictiveWarmer creates a new predictive warmer
func NewPredictiveWarmer(cache *Cache, logger log.Logger, threshold int) *PredictiveWarmer {
	return &PredictiveWarmer{
		cache:        cache,
		logger:       logger,
		accessLog:    make(map[string]*accessStats),
		warmThreshold: threshold,
	}
}

// RecordAccess records an access to a key
func (pw *PredictiveWarmer) RecordAccess(key string) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if stats, ok := pw.accessLog[key]; ok {
		stats.count++
		stats.lastAccess = time.Now()
	} else {
		pw.accessLog[key] = &accessStats{
			count:      1,
			lastAccess: time.Now(),
		}
	}
}

// GetHotKeys returns keys that exceed the access threshold
func (pw *PredictiveWarmer) GetHotKeys() []string {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	var hotKeys []string
	for key, stats := range pw.accessLog {
		if stats.count >= pw.warmThreshold {
			hotKeys = append(hotKeys, key)
		}
	}

	return hotKeys
}

// WarmHotKeys warms the hot keys
func (pw *PredictiveWarmer) WarmHotKeys(loaders map[string]func() (interface{}, error)) error {
	hotKeys := pw.GetHotKeys()

	for _, key := range hotKeys {
		if loader, ok := loaders[key]; ok {
			value, err := loader()
			if err != nil {
				pw.logger.Error("failed to load hot key",
					"key", key,
					"error", err)
				continue
			}

			if err := pw.cache.Set(key, value, pw.cache.config.DefaultTTL); err != nil {
				pw.logger.Error("failed to cache hot key",
					"key", key,
					"error", err)
			}
		}
	}

	return nil
}

// Cleanup removes stale access stats
func (pw *PredictiveWarmer) Cleanup(maxAge time.Duration) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	now := time.Now()
	for key, stats := range pw.accessLog {
		if now.Sub(stats.lastAccess) > maxAge {
			delete(pw.accessLog, key)
		}
	}
}

// ScheduledWarmer handles scheduled cache warming
type ScheduledWarmer struct {
	warmer   *Warmer
	logger   log.Logger
	schedule time.Duration
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewScheduledWarmer creates a new scheduled warmer
func NewScheduledWarmer(warmer *Warmer, logger log.Logger, schedule time.Duration) *ScheduledWarmer {
	return &ScheduledWarmer{
		warmer:   warmer,
		logger:   logger,
		schedule: schedule,
		stopChan: make(chan struct{}),
	}
}

// Start starts the scheduled warming
func (sw *ScheduledWarmer) Start(ctx context.Context, tasks []WarmupTask) {
	sw.mu.Lock()
	if sw.running {
		sw.mu.Unlock()
		return
	}
	sw.running = true
	sw.mu.Unlock()

	go sw.run(ctx, tasks)
}

// Stop stops the scheduled warming
func (sw *ScheduledWarmer) Stop() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	if !sw.running {
		return
	}

	close(sw.stopChan)
	sw.running = false
}

// run executes the scheduled warming loop
func (sw *ScheduledWarmer) run(ctx context.Context, tasks []WarmupTask) {
	ticker := time.NewTicker(sw.schedule)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sw.logger.Info("executing scheduled cache warmup")
			if err := sw.warmer.Warmup(ctx, tasks); err != nil {
				sw.logger.Error("scheduled warmup failed", "error", err)
			}
		case <-sw.stopChan:
			sw.logger.Info("scheduled warmer stopped")
			return
		case <-ctx.Done():
			sw.logger.Info("scheduled warmer context canceled")
			return
		}
	}
}
