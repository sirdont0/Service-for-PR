package pg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/you/pr-assign-avito/internal/domain"
	"github.com/you/pr-assign-avito/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ repository.Repo = (*PGRepo)(nil)

type PGRepo struct {
	pool *pgxpool.Pool
}

func NewPGRepo(pool *pgxpool.Pool) *PGRepo {
	return &PGRepo{pool: pool}
}

func (p *PGRepo) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.User) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name=$1)", teamName).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return repository.ErrTeamExists
	}

	var teamID int
	err = tx.QueryRow(ctx, "INSERT INTO teams(name) VALUES ($1) RETURNING id", teamName).Scan(&teamID)
	if err != nil {
		return err
	}

	for _, m := range members {
		_, err = tx.Exec(ctx, `INSERT INTO users (id, username, team_id, is_active) VALUES ($1,$2,$3,$4)
            ON CONFLICT (id) DO UPDATE SET username=EXCLUDED.username, team_id=EXCLUDED.team_id, is_active=EXCLUDED.is_active`, m.ID, m.Username, teamID, m.IsActive)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (p *PGRepo) GetTeamByName(ctx context.Context, name string) (domain.Team, []domain.User, error) {
	var team domain.Team
	err := p.pool.QueryRow(ctx, "SELECT id, name FROM teams WHERE name=$1", name).Scan(&team.ID, &team.Name)
	if err != nil {
		return team, nil, repository.ErrNotFound
	}
	rows, err := p.pool.Query(ctx, "SELECT id, username, team_id, is_active FROM users WHERE team_id=$1", team.ID)
	if err != nil {
		return team, nil, err
	}
	defer rows.Close()
	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive); err != nil {
			return team, nil, err
		}
		users = append(users, u)
	}
	return team, users, nil
}

func (p *PGRepo) SetUserActive(ctx context.Context, userID string, active bool) (domain.User, error) {
	tag, err := p.pool.Exec(ctx, "UPDATE users SET is_active=$1 WHERE id=$2", active, userID)
	if err != nil {
		return domain.User{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.User{}, repository.ErrNotFound
	}
	var u domain.User
	err = p.pool.QueryRow(ctx, `
        SELECT u.id, u.username, u.team_id, t.name, u.is_active
        FROM users u
        JOIN teams t ON t.id = u.team_id
        WHERE u.id=$1
    `, userID).Scan(&u.ID, &u.Username, &u.TeamID, &u.TeamName, &u.IsActive)
	if err != nil {
		return domain.User{}, repository.ErrNotFound
	}
	return u, nil
}

func (p *PGRepo) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	var u domain.User
	err := p.pool.QueryRow(ctx, `
        SELECT u.id, u.username, u.team_id, t.name, u.is_active
        FROM users u
        JOIN teams t ON t.id = u.team_id
        WHERE u.id=$1
    `, userID).Scan(&u.ID, &u.Username, &u.TeamID, &u.TeamName, &u.IsActive)
	if err != nil {
		return domain.User{}, repository.ErrNotFound
	}
	return u, nil
}

func (p *PGRepo) PRExists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id=$1)", prID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *PGRepo) CreatePR(ctx context.Context, pr domain.PullRequest, status string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var statusID int
	err = tx.QueryRow(ctx, "SELECT id FROM pr_statuses WHERE name=$1", status).Scan(&statusID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "INSERT INTO pull_requests (id, title, author_id, status_id) VALUES ($1,$2,$3,$4)",
		pr.ID, pr.Title, pr.AuthorID, statusID)
	if err != nil {
		return err
	}

	for _, r := range pr.Reviewers {
		_, err = tx.Exec(ctx, "INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1,$2)", pr.ID, r)
		if err != nil {
			return err
		}
		// Деактивируем ревьювера при назначении
		_, err = tx.Exec(ctx, "UPDATE users SET is_active=FALSE WHERE id=$1", r)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (p *PGRepo) GetPR(ctx context.Context, prID string) (domain.PullRequest, error) {
	var pr domain.PullRequest
	var statusName string
	var mergedAt pgxNullTime
	err := p.pool.QueryRow(ctx, `
        SELECT pr.id, pr.title, pr.author_id, st.name, pr.created_at, pr.merged_at
        FROM pull_requests pr
        JOIN pr_statuses st ON pr.status_id = st.id
        WHERE pr.id=$1
    `, prID).Scan(&pr.ID, &pr.Title, &pr.AuthorID, &statusName, &pr.CreatedAt, &mergedAt)
	if err != nil {
		return pr, repository.ErrNotFound
	}
	pr.Status = statusName
	if mergedAt.Valid {
		t := mergedAt.Time
		pr.MergedAt = &t
	}

	rows, err := p.pool.Query(ctx, "SELECT reviewer_id FROM pr_reviewers WHERE pr_id=$1", prID)
	if err == nil {
		defer rows.Close()
		var revs []string
		for rows.Next() {
			var rid string
			rows.Scan(&rid)
			revs = append(revs, rid)
		}
		pr.Reviewers = revs
	}
	return pr, nil
}

type pgxNullTime struct {
	Time  time.Time
	Valid bool
}

func (n *pgxNullTime) Scan(src interface{}) error {
	if src == nil {
		n.Valid = false
		return nil
	}
	n.Valid = true
	t, ok := src.(time.Time)
	if !ok {
		return fmt.Errorf("expected time.Time, got %T", src)
	}
	n.Time = t
	return nil
}

func (p *PGRepo) GetActiveTeamMembersExcluding(ctx context.Context, teamID int, exclude []string) ([]domain.User, error) {
	var users []domain.User
	q := "SELECT id, username, team_id, is_active FROM users WHERE team_id=$1 AND is_active=TRUE"
	args := []interface{}{teamID}
	if len(exclude) > 0 {
		placeholders := make([]string, len(exclude))
		for i := range exclude {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, exclude[i])
		}
		q = q + " AND id NOT IN (" + strings.Join(placeholders, ",") + ")"
	}
	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (p *PGRepo) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := p.pool.Query(ctx, "SELECT reviewer_id FROM pr_reviewers WHERE pr_id=$1", prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var revs []string
	for rows.Next() {
		var r string
		rows.Scan(&r)
		revs = append(revs, r)
	}
	return revs, nil
}

func (p *PGRepo) IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pr_id=$1 AND reviewer_id=$2)", prID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *PGRepo) GetPRAuthor(ctx context.Context, prID string) (string, error) {
	var author string
	err := p.pool.QueryRow(ctx, "SELECT author_id FROM pull_requests WHERE id=$1", prID).Scan(&author)
	if err != nil {
		return "", repository.ErrNotFound
	}
	return author, nil
}

func (p *PGRepo) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT pr.id, pr.title, pr.author_id, st.name, pr.created_at, pr.merged_at
        FROM pr_reviewers rv
        JOIN pull_requests pr ON pr.id = rv.pr_id
        JOIN pr_statuses st ON pr.status_id = st.id
        WHERE rv.reviewer_id = $1
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		var merged pgxNullTime
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &merged); err != nil {
			return nil, err
		}
		if merged.Valid {
			t := merged.Time
			pr.MergedAt = &t
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

func (p *PGRepo) ReplacePRReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `
        SELECT st.name
        FROM pull_requests pr
        JOIN pr_statuses st ON pr.status_id = st.id
        WHERE pr.id=$1
        FOR UPDATE
    `, prID).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return repository.ErrNotFound
		}
		return err
	}
	if status == "MERGED" {
		return repository.ErrPRMerged
	}

	cmd, err := tx.Exec(ctx, "DELETE FROM pr_reviewers WHERE pr_id=$1 AND reviewer_id=$2", prID, oldUserID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return repository.ErrNotAssigned
	}

	// Проверяем, есть ли у старого ревьювера другие открытые PR
	var hasOpenPRs bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 
			FROM pr_reviewers rv
			JOIN pull_requests pr ON pr.id = rv.pr_id
			JOIN pr_statuses st ON pr.status_id = st.id
			WHERE rv.reviewer_id = $1 AND st.name = 'OPEN'
		)
	`, oldUserID).Scan(&hasOpenPRs)
	if err != nil {
		return err
	}

	// Активируем старого ревьювера, если у него нет других открытых PR
	if !hasOpenPRs {
		_, err = tx.Exec(ctx, "UPDATE users SET is_active=TRUE WHERE id=$1", oldUserID)
		if err != nil {
			return err
		}
	}

	// Добавляем нового ревьювера
	if _, err := tx.Exec(ctx, "INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1,$2)", prID, newUserID); err != nil {
		return err
	}

	// Деактивируем нового ревьювера
	_, err = tx.Exec(ctx, "UPDATE users SET is_active=FALSE WHERE id=$1", newUserID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *PGRepo) MergePR(ctx context.Context, prID string) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT st.name FROM pull_requests pr JOIN pr_statuses st ON pr.status_id=st.id WHERE pr.id=$1 FOR UPDATE", prID).Scan(&status)
	if err != nil {
		return repository.ErrNotFound
	}
	if status == "MERGED" {
		return tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx, "UPDATE pull_requests SET status_id = (SELECT id FROM pr_statuses WHERE name='MERGED'), merged_at=$2 WHERE id=$1", prID, time.Now().UTC())
	if err != nil {
		return err
	}

	// Активируем всех ревьюверов после merge
	_, err = tx.Exec(ctx, `
		UPDATE users 
		SET is_active=TRUE 
		WHERE id IN (
			SELECT reviewer_id 
			FROM pr_reviewers 
			WHERE pr_id=$1
		)
	`, prID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *PGRepo) HasOpenPRsAsReviewer(ctx context.Context, userID string) (bool, error) {
	var hasOpen bool
	err := p.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 
			FROM pr_reviewers rv
			JOIN pull_requests pr ON pr.id = rv.pr_id
			JOIN pr_statuses st ON pr.status_id = st.id
			WHERE rv.reviewer_id = $1 AND st.name = 'OPEN'
		)
	`, userID).Scan(&hasOpen)
	return hasOpen, err
}
