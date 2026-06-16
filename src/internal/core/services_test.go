package core_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/store/memory"
)

// fakeHasher is a deterministic stand-in for bcrypt so unit tests stay fast and
// predictable. It is not secure and is for tests only.
type fakeHasher struct{}

func (fakeHasher) Hash(plain string) (string, error) { return "hashed:" + plain, nil }
func (fakeHasher) Compare(hash, plain string) bool   { return hash == "hashed:"+plain }

type services struct {
	users *core.UserService
	projs *core.ProjectService
	links *core.LinkService
	stats *core.StatisticService
}

func newServices() services {
	store := memory.New()
	return services{
		users: core.NewUserService(store.Users(), fakeHasher{}),
		projs: core.NewProjectService(store.Projects()),
		links: core.NewLinkService(store.Links(), store.Projects()),
		stats: core.NewStatisticService(store.Statistics(), store.Links(), store.Projects(), store.Buffer()),
	}
}

func TestUserRegisterAndAuthenticate(t *testing.T) {
	s := newServices()
	ctx := context.Background()

	u, err := s.users.Register(ctx, "Alice@Example.com", "password123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.Id == 0 {
		t.Fatal("expected an assigned id")
	}
	if u.Email != "alice@example.com" {
		t.Fatalf("email not normalised: %q", u.Email)
	}

	if _, err := s.users.Authenticate(ctx, "alice@example.com", "password123"); err != nil {
		t.Fatalf("authenticate good: %v", err)
	}

	// Wrong password and unknown email must be indistinguishable.
	if _, err := s.users.Authenticate(ctx, "alice@example.com", "wrong"); !errors.Is(err, core.ErrInvalidCredentials) {
		t.Fatalf("wrong password: want ErrInvalidCredentials, got %v", err)
	}
	if _, err := s.users.Authenticate(ctx, "nobody@example.com", "password123"); !errors.Is(err, core.ErrInvalidCredentials) {
		t.Fatalf("unknown email: want ErrInvalidCredentials, got %v", err)
	}

	// Duplicate registration conflicts (case-insensitive on email).
	if _, err := s.users.Register(ctx, "alice@example.com", "password123"); !errors.Is(err, core.ErrConflict) {
		t.Fatalf("duplicate: want ErrConflict, got %v", err)
	}
}

func TestProjectOwnershipIsolation(t *testing.T) {
	s := newServices()
	ctx := context.Background()
	alice, _ := s.users.Register(ctx, "alice@example.com", "password123")
	bob, _ := s.users.Register(ctx, "bob@example.com", "password123")

	p, err := s.projs.Create(ctx, alice.Id, "Campaign")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := s.projs.Get(ctx, alice.Id, p.Id); err != nil {
		t.Fatalf("owner get: %v", err)
	}
	// Foreign owner must see ErrNotOwner (collapsed to 404 at the HTTP edge).
	if _, err := s.projs.Get(ctx, bob.Id, p.Id); !errors.Is(err, core.ErrNotOwner) {
		t.Fatalf("foreign get: want ErrNotOwner, got %v", err)
	}

	bobList, _ := s.projs.List(ctx, bob.Id)
	if len(bobList) != 0 {
		t.Fatalf("bob should see no projects, got %d", len(bobList))
	}
	aliceList, _ := s.projs.List(ctx, alice.Id)
	if len(aliceList) != 1 {
		t.Fatalf("alice should see 1 project, got %d", len(aliceList))
	}
}

func TestLinkLifecycleAndOwnership(t *testing.T) {
	s := newServices()
	ctx := context.Background()
	alice, _ := s.users.Register(ctx, "alice@example.com", "password123")
	bob, _ := s.users.Register(ctx, "bob@example.com", "password123")
	p, _ := s.projs.Create(ctx, alice.Id, "Campaign")

	// Bob cannot create a link under Alice's project.
	if _, err := s.links.Create(ctx, bob.Id, p.Id, "Sneaky"); !errors.Is(err, core.ErrNotOwner) {
		t.Fatalf("foreign create: want ErrNotOwner, got %v", err)
	}

	l, err := s.links.Create(ctx, alice.Id, p.Id, "Newsletter")
	if err != nil {
		t.Fatalf("create link: %v", err)
	}
	if l.LinkHash == "" || !l.Active {
		t.Fatalf("link should have a hash and be active: %+v", l)
	}

	// Toggle active off, then rename.
	off := false
	if l, err = s.links.Update(ctx, alice.Id, l.Id, nil, &off); err != nil || l.Active {
		t.Fatalf("deactivate: err=%v active=%v", err, l.Active)
	}
	name := "Renamed"
	if l, err = s.links.Update(ctx, alice.Id, l.Id, &name, nil); err != nil || l.Name != "Renamed" {
		t.Fatalf("rename: err=%v name=%q", err, l.Name)
	}

	// Bob cannot read or delete it.
	if _, err := s.links.Get(ctx, bob.Id, l.Id); !errors.Is(err, core.ErrNotOwner) {
		t.Fatalf("foreign get link: want ErrNotOwner, got %v", err)
	}
	if err := s.links.Delete(ctx, alice.Id, l.Id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.links.Get(ctx, alice.Id, l.Id); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("after delete: want ErrNotFound, got %v", err)
	}
}

func TestStatisticsRecordFlushAndRead(t *testing.T) {
	s := newServices()
	ctx := context.Background()
	alice, _ := s.users.Register(ctx, "alice@example.com", "password123")
	p, _ := s.projs.Create(ctx, alice.Id, "Campaign")
	l, _ := s.links.Create(ctx, alice.Id, p.Id, "Newsletter")

	now := time.Date(2026, 6, 14, 21, 30, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		if err := s.stats.RecordHit(ctx, l.LinkHash, now); err != nil {
			t.Fatalf("record hit: %v", err)
		}
	}
	// Unknown hash and inactive link must not panic and must surface sentinels.
	if err := s.stats.RecordHit(ctx, "deadbeef", now); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("unknown hash: want ErrNotFound, got %v", err)
	}
	off := false
	_, _ = s.links.Update(ctx, alice.Id, l.Id, nil, &off)
	if err := s.stats.RecordHit(ctx, l.LinkHash, now); !errors.Is(err, core.ErrLinkInactive) {
		t.Fatalf("inactive link: want ErrLinkInactive, got %v", err)
	}

	// Hits live in the buffer until flushed.
	before, _ := s.stats.ListByLink(ctx, alice.Id, l.Id, time.Time{}, time.Time{})
	if len(before) != 0 {
		t.Fatalf("expected no persisted stats before flush, got %d", len(before))
	}
	if err := s.stats.Flush(ctx); err != nil {
		t.Fatalf("flush: %v", err)
	}
	after, _ := s.stats.ListByLink(ctx, alice.Id, l.Id, time.Time{}, time.Time{})
	if len(after) != 1 || after[0].Hits != 3 {
		t.Fatalf("want one bucket with 3 hits, got %+v", after)
	}
	if !after[0].Hour.Equal(now.Truncate(time.Hour)) {
		t.Fatalf("bucket hour: want %v, got %v", now.Truncate(time.Hour), after[0].Hour)
	}

	// Foreign owner cannot read stats.
	bob, _ := s.users.Register(ctx, "bob@example.com", "password123")
	if _, err := s.stats.ListByProject(ctx, bob.Id, p.Id, time.Time{}, time.Time{}); !errors.Is(err, core.ErrNotOwner) {
		t.Fatalf("foreign stats: want ErrNotOwner, got %v", err)
	}
}
