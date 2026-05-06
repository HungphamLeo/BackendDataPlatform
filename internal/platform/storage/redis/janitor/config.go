package janitor

import "time"

type Config struct {
	// Run cleanup every Interval. If 0 => disabled (but memory-based trigger still can run if you call Run()).
	Interval time.Duration

	// When used_memory / maxmemory >= MemoryThreshold => start aggressive cleanup.
	// If Redis maxmemory is 0 (not set), janitor will fallback to periodic cleanup only.
	MemoryThreshold float64 // e.g. 0.80

	// Only delete keys matching these patterns (safety guard).
	// Examples: "binance:ticker:*", "binance:candles:*", "binance:balance:*"
	Patterns []string

	// How many keys to scan per iteration; larger = faster but more load.
	ScanCount int64 // e.g. 1000

	// Delete in batches to avoid huge DEL commands
	DeleteBatchSize int // e.g. 500

	// Max keys to delete per run (safety)
	MaxDeletesPerRun int // e.g. 200_000
}
