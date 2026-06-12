package core

import (
	"context"
	"time"
)

type StatisticService struct {
	stats    StatisticRepository
	links    LinkRepository
	projects ProjectRepository
	buffer   HitBuffer
}

func NewStatisticService(
	stats StatisticRepository,
	links LinkRepository,
	projects ProjectRepository,
	buffer HitBuffer,
) *StatisticService {
	return &StatisticService{
		stats:    stats,
		links:    links,
		projects: projects,
		buffer:   buffer,
	}
}

// RecordHit is the pixel hot path: resolve the public hash, increment the in-memory counter
func (s *StatisticService) RecordHit(ctx context.Context, linkHash string, at time.Time) error {
	link, err := s.links.GetByHash(ctx, linkHash)
	if err != nil {
		return err
	}
	if !link.Active {
		return ErrLinkInactive
	}
	return s.buffer.Increment(ctx, link.Id, at.Truncate(time.Hour))
}

// Flush drains the hit buffer into the persistent store; only call using background worker
func (s *StatisticService) Flush(ctx context.Context) error {
	counters, err := s.buffer.Drain(ctx)
	if err != nil {
		return err
	}

	for _, c := range counters {
		if err := s.stats.UpsertHits(ctx, c.LinkId, c.Hour, c.Hits); err != nil {
			return err
		}
	}

	return nil
}

func (s *StatisticService) ListByLink(ctx context.Context, ownerId, linkId ID, from, to time.Time) ([]Statistic, error) {
	link, err := s.links.GetById(ctx, linkId)

	if err != nil {
		return nil, err
	}

	if err := s.checkProjectOwner(ctx, ownerId, link.ProjectId); err != nil {
		return nil, err
	}

	return s.stats.ListByLink(ctx, linkId, from, to)
}

func (s *StatisticService) ListByProject(ctx context.Context, ownerId, projectId ID, from, to time.Time) ([]Statistic, error) {
	if err := s.checkProjectOwner(ctx, ownerId, projectId); err != nil {
		return nil, err
	}

	return s.stats.ListByProject(ctx, projectId, from, to)
}

func (s *StatisticService) checkProjectOwner(ctx context.Context, ownerId, projectId ID) error {
	project, err := s.projects.GetById(ctx, projectId)

	if err != nil {
		return err
	}

	if project.OwnerId != ownerId {
		return ErrNotOwner
	}

	return nil
}
