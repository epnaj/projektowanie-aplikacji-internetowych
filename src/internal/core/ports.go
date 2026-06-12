package core

import (
	"context"
	"time"
)

// NOTE: All Create functions below insert the <ENTITY> and sets <ENTITY>.Id from the database

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetById(ctx context.Context, id ID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id ID) error
}

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetById(ctx context.Context, id ID) (*Project, error)
	ListByOwner(ctx context.Context, ownerId ID) ([]Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id ID) error
}

type LinkRepository interface {
	Create(ctx context.Context, link *Link) error
	GetById(ctx context.Context, id ID) (*Link, error)
	GetByHash(ctx context.Context, hash string) (*Link, error)
	ListByProject(ctx context.Context, projectId ID) ([]Link, error)
	Update(ctx context.Context, link *Link) error
	Delete(ctx context.Context, id ID) error
}

// StatisticRepository is write-only for the background worker (UpsertHits)
// and read-only for the dashboard/API; statistics are never created or
// deleted directly by users.
type StatisticRepository interface {
	UpsertHits(ctx context.Context, linkId ID, hour time.Time, hits int64) error
	ListByLink(ctx context.Context, linkId ID, from, to time.Time) ([]Statistic, error)
	ListByProject(ctx context.Context, projectId ID, from, to time.Time) ([]Statistic, error)
}

// LinkCounter is one drained counter from the hit buffer.
type LinkCounter struct {
	LinkId ID
	Hour   time.Time
	Hits   int64
}

// HitBuffer is the write-behind cache for pixel hits
type HitBuffer interface {
	Increment(ctx context.Context, linkId ID, hour time.Time) error
	// Drain atomically reads and resets all counters.
	Drain(ctx context.Context) ([]LinkCounter, error)
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) bool
}
