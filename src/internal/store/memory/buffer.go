package memory

import (
	"context"
	"sync"
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// Buffer is a standalone, mutex-guarded write-behind hit buffer. It is the hot
// path for pixel hits: many cheap Increments accumulate in memory and a
// background worker periodically Drains them into the durable StatisticRepository.
//
// It is deliberately independent of Store so it can sit in front of any backend
// (e.g. the Postgres store) without dragging the rest of the in-memory maps along.
type Buffer struct {
	mu   sync.Mutex
	hits map[statKey]int64
}

func NewBuffer() *Buffer {
	return &Buffer{hits: make(map[statKey]int64)}
}

func (b *Buffer) Increment(_ context.Context, linkId core.ID, hour time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.hits[statKey{linkId: linkId, unixHour: hour.Unix()}]++
	return nil
}

// Drain atomically reads and resets all counters.
func (b *Buffer) Drain(_ context.Context) ([]core.LinkCounter, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]core.LinkCounter, 0, len(b.hits))
	for key, hits := range b.hits {
		out = append(out, core.LinkCounter{
			LinkId: key.linkId,
			Hour:   time.Unix(key.unixHour, 0).UTC(),
			Hits:   hits,
		})
	}
	b.hits = make(map[statKey]int64)
	return out, nil
}
