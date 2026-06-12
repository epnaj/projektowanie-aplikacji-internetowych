package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"
)

type LinkService struct {
	links    LinkRepository
	projects ProjectRepository
}

func NewLinkService(links LinkRepository, projects ProjectRepository) *LinkService {
	return &LinkService{
		links:    links,
		projects: projects,
	}
}

// ownedProject loads a project and rejects (ErrNotOwner) non-owners
func (s *LinkService) ownedProject(ctx context.Context, ownerId, projectId ID) (*Project, error) {
	project, err := s.projects.GetById(ctx, projectId)

	if err != nil {
		return nil, err
	}

	if project.OwnerId != ownerId {
		return nil, ErrNotOwner
	}

	return project, nil
}

// ownedLink loads a link and verifies ownership through its project
func (s *LinkService) ownedLink(ctx context.Context, ownerId, linkId ID) (*Link, error) {
	link, err := s.links.GetById(ctx, linkId)

	if err != nil {
		return nil, err
	}

	if _, err := s.ownedProject(ctx, ownerId, link.ProjectId); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *LinkService) Create(ctx context.Context, ownerId, projectId ID, name string) (*Link, error) {
	if _, err := s.ownedProject(ctx, ownerId, projectId); err != nil {
		return nil, err
	}

	hash, err := newLinkHash()
	if err != nil {
		return nil, err
	}

	link := &Link{
		ProjectId: projectId,
		Name:      name,
		LinkHash:  hash,
		Active:    true, // should it be active by default?; for now leaving true
		CreatedAt: time.Now(),
	}
	return link, s.links.Create(ctx, link)
}

func (s *LinkService) Get(ctx context.Context, ownerId, id ID) (*Link, error) {
	return s.ownedLink(ctx, ownerId, id)
}

func (s *LinkService) ListByProject(ctx context.Context, ownerId, projectId ID) ([]Link, error) {
	if _, err := s.ownedProject(ctx, ownerId, projectId); err != nil {
		return nil, err
	}
	return s.links.ListByProject(ctx, projectId)
}

// Update changes the name and/or active flag; nil means "leave unchanged"
// consider splitting this function in the futyre
func (s *LinkService) Update(ctx context.Context, ownerId, id ID, name *string, active *bool) (*Link, error) {
	link, err := s.ownedLink(ctx, ownerId, id)

	if err != nil {
		return nil, err
	}
	if name != nil {
		link.Name = *name
	}
	if active != nil {
		link.Active = *active
	}

	return link, s.links.Update(ctx, link)
}

func (s *LinkService) Delete(ctx context.Context, ownerId, id ID) error {
	if _, err := s.ownedLink(ctx, ownerId, id); err != nil {
		return err
	}
	return s.links.Delete(ctx, id)
}

// newLinkHash generates the public token embedded in the pixel URL
func newLinkHash() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}
