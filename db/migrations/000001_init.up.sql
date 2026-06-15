CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE projects (
    id         BIGSERIAL PRIMARY KEY,
    owner_id   BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_projects_owner ON projects (owner_id);

CREATE TABLE links (
    id         BIGSERIAL PRIMARY KEY,
    project_id BIGINT      NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    link_hash  TEXT        NOT NULL UNIQUE,
    active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_links_project ON links (project_id);

-- statistics holds one row per (link, hour) bucket. The UNIQUE constraint backs
-- the write-behind upsert (ON CONFLICT) and also indexes the time-range reads.
CREATE TABLE statistics (
    id      BIGSERIAL PRIMARY KEY,
    link_id BIGINT      NOT NULL REFERENCES links (id) ON DELETE CASCADE,
    hour    TIMESTAMPTZ NOT NULL,
    hits    BIGINT      NOT NULL DEFAULT 0,
    UNIQUE (link_id, hour)
);
