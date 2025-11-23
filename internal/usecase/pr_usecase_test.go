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

func (m *memRepo) GetReviewerStats(ctx context.Context) ([]repository.ReviewerStat, error) {
	userCounts := make(map[string]int)
	for _, reviewers := range m.reviewers {
		for _, reviewerID := range reviewers {
			userCounts[reviewerID]++
		}
	}
	var stats []repository.ReviewerStat
	for userID, count := range userCounts {
		if user, ok := m.users[userID]; ok {
			stats = append(stats, repository.ReviewerStat{
				UserID:   userID,
				Username: user.Username,
				Count:    count,
			})
		}
	}
	// Добавляем пользователей без назначений
	for userID, user := range m.users {
		if _, hasAssignments := userCounts[userID]; !hasAssignments {
			stats = append(stats, repository.ReviewerStat{
				UserID:   userID,
				Username: user.Username,
				Count:    0,
			})
		}
	}
	return stats, nil
}

// Helper функции для тестов
func setupTeamWithUsers(repo *memRepo, teamName string, users []domain.User) error {
	ctx := context.Background()
	return repo.CreateTeamWithMembers(ctx, teamName, users)
}

func setupPRWithReviewers(repo *memRepo, pr domain.PullRequest, reviewers []string) {
	repo.prs[pr.ID] = pr
	repo.reviewers[pr.ID] = reviewers
}

func assertError(t *testing.T, err error, expected error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %v, got nil: %s", expected, msg)
	}
	if err != expected {
		t.Fatalf("expected %v, got %v: %s", expected, err, msg)
	}
}

func TestCreatePR_Simple(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := setupTeamWithUsers(repo, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
		{ID: "u3", Username: "carl", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
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
	if err := setupTeamWithUsers(repo, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	pr := domain.PullRequest{ID: "pr2", Title: "fix", AuthorID: "u1", Reviewers: []string{"u2"}, Status: "OPEN"}
	setupPRWithReviewers(repo, pr, []string{"u2"})
	u := NewPRUsecase(repo)
	_, err := u.ReassignReviewer(ctx, "pr2", "u2")
	assertError(t, err, ErrNoCandidate, "expected ErrNoCandidate")
}

func TestMerge_Idempotent(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
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

// Расширенные тесты для CreatePR
func TestCreatePR_PRExists(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	u := NewPRUsecase(repo)
	pr := domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "u1"}
	created, err := u.CreatePR(ctx, pr)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if created.Status != "OPEN" {
		t.Fatalf("expected OPEN got %s", created.Status)
	}
	// Пытаемся создать PR с тем же ID
	_, err = u.CreatePR(ctx, pr)
	if err == nil {
		t.Fatalf("expected ErrPRExists, got nil")
	}
	if err != ErrPRExists {
		t.Fatalf("expected ErrPRExists, got %v", err)
	}
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	u := NewPRUsecase(repo)
	pr := domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "nonexistent"}
	_, err := u.CreatePR(ctx, pr)
	if err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreatePR_NoCandidates(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	u := NewPRUsecase(repo)
	pr := domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "u1"}
	created, err := u.CreatePR(ctx, pr)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(created.Reviewers) != 0 {
		t.Fatalf("expected 0 reviewers when no candidates, got %d", len(created.Reviewers))
	}
}

func TestCreatePR_OneCandidate(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	u := NewPRUsecase(repo)
	pr := domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "u1"}
	created, err := u.CreatePR(ctx, pr)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(created.Reviewers) != 1 {
		t.Fatalf("expected 1 reviewer, got %d", len(created.Reviewers))
	}
	if created.Reviewers[0] != "u2" {
		t.Fatalf("expected reviewer u2, got %s", created.Reviewers[0])
	}
}

func TestCreatePR_MultipleCandidates(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
		{ID: "u3", Username: "carl", TeamID: 1, IsActive: true},
		{ID: "u4", Username: "dave", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	u := NewPRUsecase(repo)
	pr := domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "u1"}
	created, err := u.CreatePR(ctx, pr)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(created.Reviewers) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(created.Reviewers))
	}
	// Проверяем, что автор не в списке ревьюверов
	for _, reviewer := range created.Reviewers {
		if reviewer == "u1" {
			t.Fatalf("author should not be in reviewers list")
		}
	}
}

func TestCreatePR_ExcludesInactive(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: false},
		{ID: "u3", Username: "carl", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	u := NewPRUsecase(repo)
	pr := domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "u1"}
	created, err := u.CreatePR(ctx, pr)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// Проверяем, что неактивный пользователь не выбран
	for _, reviewer := range created.Reviewers {
		if reviewer == "u2" {
			t.Fatalf("inactive user should not be selected as reviewer")
		}
	}
}

// Расширенные тесты для ReassignReviewer
func TestReassignReviewer_Success(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
		{ID: "u3", Username: "carl", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	pr := domain.PullRequest{ID: "pr1", Title: "fix", AuthorID: "u1", Reviewers: []string{"u2"}, Status: "OPEN"}
	repo.prs[pr.ID] = pr
	repo.reviewers[pr.ID] = []string{"u2"}
	u := NewPRUsecase(repo)
	newID, err := u.ReassignReviewer(ctx, "pr1", "u2")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if newID != "u3" {
		t.Fatalf("expected new reviewer u3, got %s", newID)
	}
	// Проверяем, что новый ревьювер назначен
	reviewers, _ := repo.GetPRReviewers(ctx, "pr1")
	found := false
	for _, r := range reviewers {
		if r == "u3" {
			found = true
		}
		if r == "u2" {
			t.Fatalf("old reviewer should be removed")
		}
	}
	if !found {
		t.Fatalf("new reviewer should be assigned")
	}
}

func TestReassignReviewer_PRNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	u := NewPRUsecase(repo)
	_, err := u.ReassignReviewer(ctx, "nonexistent", "u1")
	if err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestReassignReviewer_PRAlreadyMerged(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := setupTeamWithUsers(repo, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	pr := domain.PullRequest{ID: "pr1", Title: "fix", AuthorID: "u1", Reviewers: []string{"u2"}, Status: "MERGED"}
	setupPRWithReviewers(repo, pr, []string{"u2"})
	u := NewPRUsecase(repo)
	_, err := u.ReassignReviewer(ctx, "pr1", "u2")
	assertError(t, err, ErrPRMerged, "expected ErrPRMerged")
}

func TestReassignReviewer_NotAssigned(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
		{ID: "u3", Username: "carl", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	pr := domain.PullRequest{ID: "pr1", Title: "fix", AuthorID: "u1", Reviewers: []string{"u2"}, Status: "OPEN"}
	repo.prs[pr.ID] = pr
	repo.reviewers[pr.ID] = []string{"u2"}
	u := NewPRUsecase(repo)
	_, err := u.ReassignReviewer(ctx, "pr1", "u3")
	if err == nil {
		t.Fatalf("expected ErrNotAssigned, got nil")
	}
	if err != ErrNotAssigned {
		t.Fatalf("expected ErrNotAssigned, got %v", err)
	}
}

func TestReassignReviewer_OldUserNotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := setupTeamWithUsers(repo, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	pr := domain.PullRequest{ID: "pr1", Title: "fix", AuthorID: "u1", Reviewers: []string{"u2"}, Status: "OPEN"}
	setupPRWithReviewers(repo, pr, []string{"u2"})
	u := NewPRUsecase(repo)
	// Пытаемся переназначить пользователя, который не назначен - получим ErrNotAssigned
	_, err := u.ReassignReviewer(ctx, "pr1", "nonexistent")
	assertError(t, err, ErrNotAssigned, "expected ErrNotAssigned")
}

func TestReassignReviewer_OldUserDoesNotExist(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
		{ID: "u3", Username: "carl", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	pr := domain.PullRequest{ID: "pr1", Title: "fix", AuthorID: "u1", Reviewers: []string{"u2"}, Status: "OPEN"}
	repo.prs[pr.ID] = pr
	repo.reviewers[pr.ID] = []string{"u2"}
	// Добавляем несуществующего пользователя в reviewers, чтобы пройти проверку IsReviewerAssigned
	// но затем GetUserByID вернет ошибку
	repo.reviewers["pr1"] = []string{"nonexistent"}
	u := NewPRUsecase(repo)
	_, err := u.ReassignReviewer(ctx, "pr1", "nonexistent")
	if err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// Расширенные тесты для MergePR
func TestMergePR_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	u := NewPRUsecase(repo)
	_, err := u.MergePR(ctx, "nonexistent")
	if err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMergePR_Success(t *testing.T) {
	ctx := context.Background()
	repo := newMemRepo()
	if err := repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
		{ID: "u1", Username: "alice", TeamID: 1, IsActive: true},
		{ID: "u2", Username: "bob", TeamID: 1, IsActive: true},
	}); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	pr := domain.PullRequest{ID: "pr1", Title: "merge", AuthorID: "u1", Status: "OPEN", Reviewers: []string{"u2"}}
	repo.prs[pr.ID] = pr
	repo.reviewers[pr.ID] = []string{"u2"}
	u := NewPRUsecase(repo)
	merged, err := u.MergePR(ctx, "pr1")
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if merged.Status != "MERGED" {
		t.Fatalf("expected MERGED got %s", merged.Status)
	}
	if merged.MergedAt == nil {
		t.Fatalf("expected MergedAt to be set")
	}
}

// Тесты для вспомогательных функций
func TestPickUpTo(t *testing.T) {
	tests := []struct {
		name     string
		ids      []string
		n        int
		expected int
	}{
		{"n greater than length", []string{"a", "b"}, 5, 2},
		{"n equal to length", []string{"a", "b"}, 2, 2},
		{"n less than length", []string{"a", "b", "c", "d"}, 2, 2},
		{"empty slice", []string{}, 2, 0},
		{"n is zero", []string{"a", "b"}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pickUpTo(tt.ids, tt.n)
			if len(result) != tt.expected {
				t.Fatalf("expected length %d, got %d", tt.expected, len(result))
			}
			// Проверяем, что элементы те же
			for i := 0; i < len(result) && i < len(tt.ids); i++ {
				if result[i] != tt.ids[i] {
					t.Fatalf("expected element %s at index %d, got %s", tt.ids[i], i, result[i])
				}
			}
		})
	}
}
