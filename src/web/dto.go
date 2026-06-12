package web

import (
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// data transfer objects
// yet to be used

// Auth

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type userResponse struct {
	Id        core.ID   `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func toUserResponse(u *core.User) userResponse {
	return userResponse{Id: u.Id, Email: u.Email, CreatedAt: u.CreatedAt}
}

// Project

type createProjectRequest struct {
	Name string `json:"name" validate:"required,min=2,max=80"`
}

type updateProjectRequest struct {
	Name string `json:"name" validate:"required,min=2,max=80"`
}

type projectResponse struct {
	Id        core.ID   `json:"id"`
	OwnerId   core.ID   `json:"ownerId"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func toProjectResponse(p *core.Project) projectResponse {
	return projectResponse{Id: p.Id, OwnerId: p.OwnerId, Name: p.Name, CreatedAt: p.CreatedAt}
}

func toProjectResponses(projects []core.Project) []projectResponse {
	out := make([]projectResponse, len(projects))
	for i := range projects {
		out[i] = toProjectResponse(&projects[i])
	}
	return out
}

// Link

type createLinkRequest struct {
	Name string `json:"name" validate:"required,min=2,max=80"`
}

type updateLinkRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=2,max=80"`
	Active *bool   `json:"active,omitempty"`
}

type linkResponse struct {
	Id        core.ID   `json:"id"`
	ProjectId core.ID   `json:"projectId"`
	Name      string    `json:"name"`
	LinkHash  string    `json:"linkHash"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

func toLinkResponse(l *core.Link) linkResponse {
	return linkResponse{
		Id:        l.Id,
		ProjectId: l.ProjectId,
		Name:      l.Name,
		LinkHash:  l.LinkHash,
		Active:    l.Active,
		CreatedAt: l.CreatedAt,
	}
}

func toLinkResponses(links []core.Link) []linkResponse {
	out := make([]linkResponse, len(links))
	for i := range links {
		out[i] = toLinkResponse(&links[i])
	}
	return out
}

// Statistic

type statisticResponse struct {
	LinkId core.ID   `json:"linkId"`
	Hour   time.Time `json:"hour"`
	Hits   int64     `json:"hits"`
}

func toStatisticResponse(s *core.Statistic) statisticResponse {
	return statisticResponse{
		LinkId: s.LinkId,
		Hour:   s.Hour,
		Hits:   s.Hits,
	}
}

func toStatisticResponses(stats []core.Statistic) []statisticResponse {
	output := make([]statisticResponse, len(stats))
	for i := range stats {
		output[i] = toStatisticResponse(&stats[i])
	}
	return output
}
