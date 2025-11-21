package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/you/pr-assign-avito/internal/domain"
	"github.com/you/pr-assign-avito/internal/repository"
)

type memRepo struct {
	teams     map[string]int
	users     map[string]domain.User
	prs       map[string]domain.PullRequest
	reviewers map[string][]string
	statuses  map[string]int
}

func newMemRepo() *memRepo {
	m := &memRepo{
		teams:     map[string]int{},
		users:     map[string]domain.User{},
		prs:       map[string]domain.PullRequest{},
		reviewers: map[string][]string{},
		statuses:  map[string]int{"OPEN": 1, "MERGED": 2},
	}
	return m
}

func (m *memRepo) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.User) error {
	if _, exists := m.teams[teamName]; exists {
		return repository.ErrTeamExists
	}
	id := len(m.teams) + 1
	m.teams[teamName] = id
	for _, u := range members {
		u.TeamID = id
		u.TeamName = teamName
		m.users[u.ID] = u
	}
	return nil
}
func (m *memRepo) GetTeamByName(ctx context.Context, name string) (domain.Team, []domain.User, error) {
	id, ok := m.teams[name]
	if !ok {
		return domain.Team{}, nil, repository.ErrNotFound
	}
	var list []domain.User
	for _, u := range m.users {
		if u.TeamID == id {
			list = append(list, u)
		}
	}
	return domain.Team{ID: id, Name: name}, list, nil
}
func (m *memRepo) SetUserActive(ctx context.Context, userID string, active bool) (domain.User, error) {
	u, ok := m.users[userID]
	if !ok {
		return domain.User{}, repository.ErrNotFound
	}
	u.IsActive = active
	m.users[userID] = u
	u.TeamName = m.teamNameByID(u.TeamID)
	return u, nil
}
func (m *memRepo) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	u, ok := m.users[userID]
	if !ok {
		return domain.User{}, repository.ErrNotFound
	}
	u.TeamName = m.teamNameByID(u.TeamID)
	return u, nil
}
func (m *memRepo) PRExists(ctx context.Context, prID string) (bool, error) {
	_, ok := m.prs[prID]
	return ok, nil
}
func (m *memRepo) CreatePR(ctx context.Context, pr domain.PullRequest, status string) error {
	m.prs[pr.ID] = pr
	m.reviewers[pr.ID] = append([]string{}, pr.Reviewers...)
	return nil
}
func (m *memRepo) GetActiveTeamMembersExcluding(ctx context.Context, teamID int, exclude []string) ([]domain.User, error) {
	set := map[string]struct{}{}
	for _, e := range exclude {
		set[e] = struct{}{}
	}
	var res []domain.User
	for _, u := range m.users {
		if u.TeamID == teamID && u.IsActive {
			if _, ex := set[u.ID]; !ex {
				res = append(res, u)
			}
		}
	}
	return res, nil
}
func (m *memRepo) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
	return m.reviewers[prID], nil
}
func (m *memRepo) ReplacePRReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	pr, ok := m.prs[prID]
	if !ok {
		return repository.ErrNotFound
	}
	if pr.Status == "MERGED" {
		return repository.ErrPRMerged
	}
	arr := m.reviewers[prID]
	found := false
	for i, v := range arr {
		if v == oldUserID {
			arr[i] = newUserID
			found = true
			break
		}
	}
	if !found {
		return repository.ErrNotAssigned
	}
	m.reviewers[prID] = arr
	pr.Reviewers = append([]string{}, arr...)
	m.prs[prID] = pr
	return nil
}
func (m *memRepo) MergePR(ctx context.Context, prID string) error {
	pr, ok := m.prs[prID]
	if !ok {
		return repository.ErrNotFound
	}
	if pr.Status == "MERGED" {
		return nil
	}
	pr.Status = "MERGED"
	t := time.Now().UTC()
	pr.MergedAt = &t
	m.prs[prID] = pr
	return nil
}
func (m *memRepo) GetPR(ctx context.Context, prID string) (domain.PullRequest, error) {
	pr, ok := m.prs[prID]
	if !ok {
		return domain.PullRequest{}, repository.ErrNotFound
	}
	return pr, nil
}
func (m *memRepo) IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	for _, r := range m.reviewers[prID] {
		if r == userID {
			return true, nil
		}
	}
	return false, nil
}
func (m *memRepo) GetPRAuthor(ctx context.Context, prID string) (string, error) {
	pr, ok := m.prs[prID]
	if !ok {
		return "", repository.ErrNotFound
	}
	return pr.AuthorID, nil
}
func (m *memRepo) HasOpenPRsAsReviewer(ctx context.Context, userID string) (bool, error) {
	for prID, revs := range m.reviewers {
		for _, r := range revs {
			if r == userID {
				if pr, ok := m.prs[prID]; ok && pr.Status == "OPEN" {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
func (m *memRepo) LockPRForUpdate(ctx context.Context, prID string) (string, error) {
	pr, ok := m.prs[prID]
	if !ok {
		return "", repository.ErrNotFound
	}
	return pr.Status, nil
}
func (m *memRepo) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	var res []domain.PullRequest
	for prID, revs := range m.reviewers {
		for _, r := range revs {
			if r == userID {
				if pr, ok := m.prs[prID]; ok {
					res = append(res, pr)
				}
				break
			}
		}
	}
	return res, nil
}

func (m *memRepo) teamNameByID(id int) string {
	for name, storedID := range m.teams {
		if storedID == id {
			return name
		}
	}
	return ""
}

func TestCreatePR_Simple(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
		{ID: "u3", Username: "carl", TeamID: 1, IsActive: true},
	})
	u := NewPRUsecase(repo)
	pr := domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "u1"}
	created, err := u.CreatePR(ctx, pr)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if created.Status != "OPEN" {
		t.Fatalf("expected OPEN got %s", created.Status)
	}
	if len(created.Reviewers) == 0 {
		t.Fatalf("expected reviewers assigned")
	}
}

func TestReassign_NoCandidate(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
	})
	pr := domain.PullRequest{ID: "pr2", Title: "fix", AuthorID: "u1", Reviewers: []string{"u2"}, Status: "OPEN"}
	repo.prs[pr.ID] = pr
	repo.reviewers[pr.ID] = []string{"u2"}
	u := NewPRUsecase(repo)
	_, err := u.ReassignReviewer(ctx, "pr2", "u2")
	if err == nil {
		t.Fatalf("expected NO_CANDIDATE, got nil")
	}
	if err != ErrNoCandidate {
		t.Fatalf("expected ErrNoCandidate, got %v", err)
	}
}

func TestMerge_Idempotent(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
	})
	pr := domain.PullRequest{ID: "pr3", Title: "merge", AuthorID: "u1", Status: "OPEN"}
	repo.prs[pr.ID] = pr
	u := NewPRUsecase(repo)
	_, err := u.MergePR(ctx, "pr3")
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	_, err = u.MergePR(ctx, "pr3")
	if err != nil {
		t.Fatalf("second merge failed: %v", err)
	}
	got, _ := repo.GetPR(ctx, "pr3")
	if got.Status != "MERGED" {
		t.Fatalf("expected MERGED got %s", got.Status)
	}
}
