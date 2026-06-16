# seed — simulated clients for demo data

A standalone tool that fills a running pixel-tracker instance with demo data. It is an external black box client.

Each simulated user runs concurrently with its own cookie jar: it registers,
logs in, creates a few projects and tracking links, and then fetches the
tracking pixel many times to generate statistics. Re-running tops up data
(existing accounts are reused, not duplicated).

## Run

The app must already be up (`docker compose up --build` from the repo root).

```sh
cd seed
go run . -url http://localhost:8080
```

Then log in at `http://localhost:8080/login` with any seeded account, e.g.
`seed-user-01@example.test` / `password123`.

## Flags

| Flag             | Default                 | Meaning                                   |
|------------------|-------------------------|-------------------------------------------|
| `-url`           | `http://localhost:8080` | base URL of the running app (`SEED_URL`)  |
| `-users`         | `5`                     | number of simulated accounts              |
| `-projects`      | `3`                     | max projects per user (1..N, random)      |
| `-links`         | `4`                     | max links per project (1..N, random)      |
| `-min-hits`      | `5`                     | min pixel hits per link                   |
| `-max-hits`      | `60`                    | max pixel hits per link                   |
| `-inactive-rate` | `0.1`                   | fraction of links to deactivate           |
| `-concurrency`   | `4`                     | users simulated in parallel               |
| `-password`      | `password123`           | shared password for seeded accounts       |
| `-seed`          | time-based              | RNG seed for reproducible runs            |
| `-timeout`       | `10s`                   | per-request timeout                       |
