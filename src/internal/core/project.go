package core

import (
	"context"
	"time"
)

type ProjectService struct {
	projects ProjectRepository
}

func NewProjectService(projects ProjectRepository) *ProjectService {
	return &ProjectService{projects: projects}
}

func (s *ProjectService) Create(ctx context.Context, ownerId ID, name string) (*Project, error) {
	project := &Project{
		OwnerId:   ownerId,
		Name:      name,
		CreatedAt: time.Now(),
	}
	return project, s.projects.Create(ctx, project)
}

// Get returns ErrNotOwner when the project belongs to another owner.
func (s *ProjectService) Get(ctx context.Context, ownerId, id ID) (*Project, error) {
	project, err := s.projects.GetById(ctx, id)

	if err != nil {
		return nil, err
	}

	if project.OwnerId != ownerId {
		return nil, ErrNotOwner
	}

	return project, nil
}

func (s *ProjectService) List(ctx context.Context, ownerId ID) ([]Project, error) {
	return s.projects.ListByOwner(ctx, ownerId)
}

func (s *ProjectService) Rename(ctx context.Context, ownerId, id ID, name string) (*Project, error) {
	project, err := s.Get(ctx, ownerId, id)

	if err != nil {
		return nil, err
	}
	project.Name = name

	return project, s.projects.Update(ctx, project)
}

func (s *ProjectService) Delete(ctx context.Context, ownerId, id ID) error {
	if _, err := s.Get(ctx, ownerId, id); err != nil {
		return err
	}
	return s.projects.Delete(ctx, id)
}
