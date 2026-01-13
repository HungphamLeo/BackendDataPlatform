package janitor

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	redisclient "github.com/HungphamLeo/BackendDataPlatform/internal/platform/storage/redis"
)

type Janitor struct {
	rdb  *redisclient.RedisClient
	conf Config
}

func New(rdb *redisclient.RedisClient, conf Config) *Janitor {
	if conf.ScanCount <= 0 {
		conf.ScanCount = 1000
	}
	if conf.DeleteBatchSize <= 0 {
		conf.DeleteBatchSize = 500
	}
	if conf.MaxDeletesPerRun <= 0 {
		conf.MaxDeletesPerRun = 200000
	}
	if conf.MemoryThreshold <= 0 {
		conf.MemoryThreshold = 0.80
	}
	return &Janitor{rdb: rdb, conf: conf}
}

func (j *Janitor) Run(ctx context.Context) error {
	// initial ping
	_ = j.rdb.Ping(ctx)

	// if no interval set, still allow manual / external scheduler call to CleanupOnce
	if j.conf.Interval <= 0 {
		_, _ = j.CleanupOnce(ctx)
		return nil
	}

	t := time.NewTicker(j.conf.Interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			_, _ = j.CleanupOnce(ctx)
		}
	}
}

func (j *Janitor) CleanupOnce(ctx context.Context) (int, error) {
	if len(j.conf.Patterns) == 0 {
		return 0, fmt.Errorf("janitor: no Patterns configured (refuse to delete everything)")
	}

	// Check memory pressure first; if cannot determine => do periodic cleanup anyway
	pressure, ok := j.memoryPressure(ctx)

	deleted := 0
	for _, pattern := range j.conf.Patterns {
		// If not under pressure, do a “light cleanup” (still bounded)
		// If under pressure, do “aggressive cleanup” (same logic, just continue across patterns)
		n, err := j.deleteByPattern(ctx, pattern, j.conf.MaxDeletesPerRun-deleted)
		deleted += n
		if err != nil {
			// continue next pattern, but report error if nothing deleted
			if deleted == 0 {
				return deleted, err
			}
		}
		if deleted >= j.conf.MaxDeletesPerRun {
			break
		}

		// Under heavy pressure you may want to stop after first patterns or keep going.
		if ok && pressure < j.conf.MemoryThreshold {
			// not pressured: only light pass across all patterns
			continue
		}
	}

	return deleted, nil
}

// deleteByPattern scans keys and deletes them in batches.
func (j *Janitor) deleteByPattern(ctx context.Context, pattern string, remainingBudget int) (int, error) {
	if remainingBudget <= 0 {
		return 0, nil
	}
	it := j.rdb.ScanKeys(ctx, pattern, j.conf.ScanCount)

	batch := make([]string, 0, j.conf.DeleteBatchSize)
	deleted := 0

	for it.Next(ctx) {
		batch = append(batch, it.Val())
		if len(batch) >= j.conf.DeleteBatchSize {
			if err := j.rdb.Del(ctx, batch...); err != nil {
				return deleted, err
			}
			deleted += len(batch)
			batch = batch[:0]
			if deleted >= remainingBudget {
				break
			}
		}
	}
	if err := it.Err(); err != nil {
		return deleted, err
	}
	if len(batch) > 0 && deleted < remainingBudget {
		if err := j.rdb.Del(ctx, batch...); err != nil {
			return deleted, err
		}
		deleted += len(batch)
	}
	return deleted, nil
}

// memoryPressure tries to compute used/maxmemory ratio.
// Returns (ratio, true) if maxmemory is set; otherwise (_, false).
func (j *Janitor) memoryPressure(ctx context.Context) (float64, bool) {
	info, err := j.rdb.InfoMemory(ctx)
	if err != nil {
		return 0, false
	}

	var used, max int64
	lines := strings.Split(info, "\n")
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "used_memory:") {
			used = parseInt(strings.TrimPrefix(ln, "used_memory:"))
		}
		if strings.HasPrefix(ln, "maxmemory:") {
			max = parseInt(strings.TrimPrefix(ln, "maxmemory:"))
		}
	}
	if max <= 0 {
		return 0, false
	}
	return float64(used) / float64(max), true
}

func parseInt(s string) int64 {
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
