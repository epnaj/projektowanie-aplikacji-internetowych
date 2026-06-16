// Command seed populates a running pixel-tracker instance with demo data by
// driving it the way real users and third-party pages would: it registers
// accounts, logs in, creates projects and tracking links, and fetches the
// tracking pixel many times to generate statistics. It is a standalone
// black-box client and never imports the server's code.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type config struct {
	url         string
	users       int
	maxProjects int
	maxLinks    int
	minHits     int
	maxHits     int
	concurrency int
	password    string
	inactive    float64
	timeout     time.Duration
	seed        uint64
}

var (
	projectNames = []string{
		"Marketing Site", "Product Launch", "Q3 Campaign", "Newsletter Funnel",
		"Docs Portal", "Beta Program", "Affiliate Push", "Holiday Sale",
	}
	linkNames = []string{
		"Homepage", "Pricing", "Blog Post", "Docs", "Signup", "Landing Page",
		"Newsletter", "Contact", "Features", "Changelog",
	}
)

type totals struct {
	users    atomic.Int64
	reused   atomic.Int64
	projects atomic.Int64
	links    atomic.Int64
	inactive atomic.Int64
	hits     atomic.Int64
}

func main() {
	cfg := parseFlags()
	log.SetFlags(0)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("seeding %s with %d simulated users (concurrency %d)\n", cfg.url, cfg.users, cfg.concurrency)

	var tot totals
	var wg sync.WaitGroup
	slots := make(chan struct{}, cfg.concurrency)

	for i := range cfg.users {
		wg.Add(1)
		slots <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-slots }()

			email := fmt.Sprintf("seed-user-%02d@example.test", idx+1)
			// Each user gets an independent RNG so concurrent runs stay
			// deterministic for a given -seed.
			rng := rand.New(rand.NewPCG(cfg.seed, uint64(idx)+1))

			if err := runUser(ctx, cfg, rng, email, &tot); err != nil && ctx.Err() == nil {
				log.Printf("user %s: %v", email, err)
			}
		}(i)
	}

	wg.Wait()

	if ctx.Err() != nil {
		log.Println("interrupted; partial data may have been created")
	}
	printSummary(cfg, &tot)
}

// runUser plays out one user's whole session against the public API.
func runUser(ctx context.Context, cfg config, rng *rand.Rand, email string, tot *totals) error {
	client, err := NewClient(cfg.url, cfg.timeout)
	if err != nil {
		return err
	}

	switch err := client.Register(ctx, email, cfg.password); {
	case err == nil:
		tot.users.Add(1)
	case err == ErrExists:
		tot.reused.Add(1)
	default:
		return fmt.Errorf("register: %w", err)
	}

	if err := client.Login(ctx, email, cfg.password); err != nil {
		return fmt.Errorf("login: %w", err)
	}

	nProjects := 1 + rng.IntN(cfg.maxProjects)
	for p := range nProjects {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		name := fmt.Sprintf("%s %d", pick(rng, projectNames), p+1)
		projectID, err := client.CreateProject(ctx, name)
		if err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		tot.projects.Add(1)

		nLinks := 1 + rng.IntN(cfg.maxLinks)
		for l := range nLinks {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			linkName := fmt.Sprintf("%s %d", pick(rng, linkNames), l+1)
			linkID, hash, err := client.CreateLink(ctx, projectID, linkName)
			if err != nil {
				return fmt.Errorf("create link: %w", err)
			}
			tot.links.Add(1)

			// Deactivate a fraction of links to exercise the always-200 pixel
			// on inactive links (which records no hit).
			active := true
			if rng.Float64() < cfg.inactive {
				if err := client.SetLinkActive(ctx, linkID, false); err != nil {
					return fmt.Errorf("deactivate link: %w", err)
				}
				active = false
				tot.inactive.Add(1)
			}

			hits := cfg.minHits + rng.IntN(cfg.maxHits-cfg.minHits+1)
			for range hits {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if err := client.Hit(ctx, hash); err != nil {
					return fmt.Errorf("hit pixel: %w", err)
				}
				if active {
					tot.hits.Add(1)
				}
			}
		}
	}
	return nil
}

func printSummary(cfg config, tot *totals) {
	log.Println()
	log.Println("seed complete")
	log.Printf("  accounts created : %d", tot.users.Load())
	log.Printf("  accounts reused  : %d", tot.reused.Load())
	log.Printf("  projects         : %d", tot.projects.Load())
	log.Printf("  links            : %d (%d deactivated)", tot.links.Load(), tot.inactive.Load())
	log.Printf("  pixel hits sent  : counted on active links only -> %d recorded", tot.hits.Load())
	log.Println()
	log.Printf("log in at %s/login with any seeded account, e.g.:", cfg.url)
	log.Printf("  email:    seed-user-01@example.test")
	log.Printf("  password: %s", cfg.password)
	log.Println()
	log.Println("note: hits are timestamped server-side, so they land in the current hour bucket.")
}

func pick(rng *rand.Rand, options []string) string {
	return options[rng.IntN(len(options))]
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.url, "url", envOr("SEED_URL", "http://localhost:8080"), "base URL of the running app")
	flag.IntVar(&cfg.users, "users", 5, "number of simulated user accounts")
	flag.IntVar(&cfg.maxProjects, "projects", 3, "max projects per user (1..N, random)")
	flag.IntVar(&cfg.maxLinks, "links", 4, "max links per project (1..N, random)")
	flag.IntVar(&cfg.minHits, "min-hits", 5, "minimum pixel hits per link")
	flag.IntVar(&cfg.maxHits, "max-hits", 60, "maximum pixel hits per link")
	flag.IntVar(&cfg.concurrency, "concurrency", 4, "number of users to simulate in parallel")
	flag.StringVar(&cfg.password, "password", "password123", "shared password for seeded accounts (min 8 chars)")
	flag.Float64Var(&cfg.inactive, "inactive-rate", 0.1, "fraction of links to deactivate (0..1)")
	flag.DurationVar(&cfg.timeout, "timeout", 10*time.Second, "per-request timeout")
	flag.Uint64Var(&cfg.seed, "seed", uint64(time.Now().UnixNano()), "RNG seed for reproducible runs")
	flag.Parse()

	if cfg.users < 1 || cfg.maxProjects < 1 || cfg.maxLinks < 1 {
		log.Fatal("users, projects and links must each be >= 1")
	}
	if cfg.minHits < 0 || cfg.maxHits < cfg.minHits {
		log.Fatal("require 0 <= min-hits <= max-hits")
	}
	if len(cfg.password) < 8 {
		log.Fatal("password must be at least 8 characters")
	}
	if cfg.concurrency < 1 {
		cfg.concurrency = 1
	}
	return cfg
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
