DROP TRIGGER IF EXISTS pr_reviewers_limit ON pr_reviewers;
DROP FUNCTION IF EXISTS check_pr_reviewer_limit;
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS pr_statuses;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
