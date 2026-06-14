package memory

import (
	"context"
	"sync"
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// statKey identifies a single (link, hour) statistics bucket. time.Time is
// avoided as a map key because of its monotonic/location components; the
// truncated hour is stored as a Unix second instead.
type statKey struct {
	linkId   core.ID
	unixHour int64
}

// Store is the shared backing data. All repositories returned by its
// accessors share this one mutex, so cross-resource reads (e.g. listing a
// project's statistics through its links) stay consistent
type Store struct {
	mu       sync.Mutex
	nextID   core.ID
	users    map[core.ID]core.User
	projects map[core.ID]core.Project
	links    map[core.ID]core.Link
	stats    map[statKey]core.Statistic
	buffer   map[statKey]int64
}

func New() *Store {
	return &Store{
		nextID:   1,
		users:    make(map[core.ID]core.User),
		projects: make(map[core.ID]core.Project),
		links:    make(map[core.ID]core.Link),
		stats:    make(map[statKey]core.Statistic),
		buffer:   make(map[statKey]int64),
	}
}

// Repository accessors. Each returns a thin view that satisfies one core
// interface; they all share s.mu via the embedded *Store.
func (s *Store) Users() core.UserRepository           { return userRepo{s} }
func (s *Store) Projects() core.ProjectRepository     { return projectRepo{s} }
func (s *Store) Links() core.LinkRepository           { return linkRepo{s} }
func (s *Store) Statistics() core.StatisticRepository { return statRepo{s} }
func (s *Store) Buffer() core.HitBuffer               { return hitBuffer{s} }

// Users

type userRepo struct{ s *Store }

func (r userRepo) Create(_ context.Context, user *core.User) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	user.Id = r.s.nextID
	r.s.nextID++
	r.s.users[user.Id] = *user
	return nil
}

func (r userRepo) GetById(_ context.Context, id core.ID) (*core.User, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	u, ok := r.s.users[id]
	if !ok {
		return nil, core.ErrNotFound
	}
	return &u, nil
}

func (r userRepo) GetByEmail(_ context.Context, email string) (*core.User, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, u := range r.s.users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, core.ErrNotFound
}

func (r userRepo) Update(_ context.Context, user *core.User) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.users[user.Id]; !ok {
		return core.ErrNotFound
	}
	r.s.users[user.Id] = *user
	return nil
}

func (r userRepo) Delete(_ context.Context, id core.ID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.users[id]; !ok {
		return core.ErrNotFound
	}
	delete(r.s.users, id)
	return nil
}

// Projects

type projectRepo struct{ s *Store }

func (r projectRepo) Create(_ context.Context, project *core.Project) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	project.Id = r.s.nextID
	r.s.nextID++
	r.s.projects[project.Id] = *project
	return nil
}

func (r projectRepo) GetById(_ context.Context, id core.ID) (*core.Project, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	p, ok := r.s.projects[id]
	if !ok {
		return nil, core.ErrNotFound
	}
	return &p, nil
}

func (r projectRepo) ListByOwner(_ context.Context, ownerId core.ID) ([]core.Project, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	out := []core.Project{}
	for _, p := range r.s.projects {
		if p.OwnerId == ownerId {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r projectRepo) Update(_ context.Context, project *core.Project) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.projects[project.Id]; !ok {
		return core.ErrNotFound
	}
	r.s.projects[project.Id] = *project
	return nil
}

func (r projectRepo) Delete(_ context.Context, id core.ID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.projects[id]; !ok {
		return core.ErrNotFound
	}
	delete(r.s.projects, id)
	return nil
}

// Links

type linkRepo struct{ s *Store }

func (r linkRepo) Create(_ context.Context, link *core.Link) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	link.Id = r.s.nextID
	r.s.nextID++
	r.s.links[link.Id] = *link
	return nil
}

func (r linkRepo) GetById(_ context.Context, id core.ID) (*core.Link, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	l, ok := r.s.links[id]
	if !ok {
		return nil, core.ErrNotFound
	}
	return &l, nil
}

func (r linkRepo) GetByHash(_ context.Context, hash string) (*core.Link, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for _, l := range r.s.links {
		if l.LinkHash == hash {
			return &l, nil
		}
	}
	return nil, core.ErrNotFound
}

func (r linkRepo) ListByProject(_ context.Context, projectId core.ID) ([]core.Link, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	out := []core.Link{}
	for _, l := range r.s.links {
		if l.ProjectId == projectId {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r linkRepo) Update(_ context.Context, link *core.Link) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.links[link.Id]; !ok {
		return core.ErrNotFound
	}
	r.s.links[link.Id] = *link
	return nil
}

func (r linkRepo) Delete(_ context.Context, id core.ID) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.links[id]; !ok {
		return core.ErrNotFound
	}
	delete(r.s.links, id)
	return nil
}

// Statistics

type statRepo struct{ s *Store }

func (r statRepo) UpsertHits(_ context.Context, linkId core.ID, hour time.Time, hits int64) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	key := statKey{linkId: linkId, unixHour: hour.Unix()}
	if existing, ok := r.s.stats[key]; ok {
		existing.Hits += hits
		r.s.stats[key] = existing
		return nil
	}
	r.s.stats[key] = core.Statistic{
		Id:     r.s.nextID,
		LinkId: linkId,
		Hour:   hour,
		Hits:   hits,
	}
	r.s.nextID++
	return nil
}

func (r statRepo) ListByLink(_ context.Context, linkId core.ID, from, to time.Time) ([]core.Statistic, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	out := []core.Statistic{}
	for _, st := range r.s.stats {
		if st.LinkId == linkId && inRange(st.Hour, from, to) {
			out = append(out, st)
		}
	}
	return out, nil
}

func (r statRepo) ListByProject(_ context.Context, projectId core.ID, from, to time.Time) ([]core.Statistic, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	linkIds := map[core.ID]bool{}
	for _, l := range r.s.links {
		if l.ProjectId == projectId {
			linkIds[l.Id] = true
		}
	}
	out := []core.Statistic{}
	for _, st := range r.s.stats {
		if linkIds[st.LinkId] && inRange(st.Hour, from, to) {
			out = append(out, st)
		}
	}
	return out, nil
}

// inRange treats the window as [from, to); a zero bound means "unbounded".
func inRange(t, from, to time.Time) bool {
	if !from.IsZero() && t.Before(from) {
		return false
	}
	if !to.IsZero() && !t.Before(to) {
		return false
	}
	return true
}

// Hit buffer (write-behind cache)

type hitBuffer struct{ s *Store }

func (b hitBuffer) Increment(_ context.Context, linkId core.ID, hour time.Time) error {
	b.s.mu.Lock()
	defer b.s.mu.Unlock()
	b.s.buffer[statKey{linkId: linkId, unixHour: hour.Unix()}]++
	return nil
}

func (b hitBuffer) Drain(_ context.Context) ([]core.LinkCounter, error) {
	b.s.mu.Lock()
	defer b.s.mu.Unlock()
	out := make([]core.LinkCounter, 0, len(b.s.buffer))
	for key, hits := range b.s.buffer {
		out = append(out, core.LinkCounter{
			LinkId: key.linkId,
			Hour:   time.Unix(key.unixHour, 0).UTC(),
			Hits:   hits,
		})
	}
	b.s.buffer = make(map[statKey]int64)
	return out, nil
}
