-- teams, users, pr_statuses, pull_requests, pr_reviewers
CREATE TABLE IF NOT EXISTS teams (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY, -- external id like U1
  username TEXT NOT NULL,
  team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_id, is_active);

CREATE TABLE IF NOT EXISTS pr_statuses (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
);
-- insert statuses
INSERT INTO pr_statuses (name) VALUES ('OPEN') ON CONFLICT DO NOTHING;
INSERT INTO pr_statuses (name) VALUES ('MERGED') ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS pull_requests (
  id TEXT PRIMARY KEY,  -- external id like PR42
  title TEXT NOT NULL,
  author_id TEXT NOT NULL REFERENCES users(id),
  status_id INTEGER NOT NULL REFERENCES pr_statuses(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  merged_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests (author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests (status_id);

CREATE TABLE IF NOT EXISTS pr_reviewers (
  pr_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
  reviewer_id TEXT NOT NULL REFERENCES users(id),
  PRIMARY KEY (pr_id, reviewer_id)
);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers (reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr ON pr_reviewers (pr_id);
