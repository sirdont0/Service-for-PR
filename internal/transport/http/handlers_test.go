package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/you/pr-assign-avito/internal/domain"
	"github.com/you/pr-assign-avito/internal/infra"
	"github.com/you/pr-assign-avito/internal/repository"
	uc "github.com/you/pr-assign-avito/internal/usecase"
)

// Mock repository для тестов handlers
type mockRepo struct {
	teams     map[string]domain.Team
	users     map[string]domain.User
	prs       map[string]domain.PullRequest
	reviewers map[string][]string
	stats     []repository.ReviewerStat
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		teams:     make(map[string]domain.Team),
		users:     make(map[string]domain.User),
		prs:       make(map[string]domain.PullRequest),
		reviewers: make(map[string][]string),
		stats:     make([]repository.ReviewerStat, 0),
	}
}

func (m *mockRepo) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.User) error {
	if _, exists := m.teams[teamName]; exists {
		return repository.ErrTeamExists
	}
	m.teams[teamName] = domain.Team{ID: len(m.teams) + 1, Name: teamName}
	for _, u := range members {
		m.users[u.ID] = u
	}
	return nil
}

func (m *mockRepo) GetTeamByName(ctx context.Context, name string) (domain.Team, []domain.User, error) {
	team, ok := m.teams[name]
	if !ok {
		return domain.Team{}, nil, repository.ErrNotFound
	}
	var members []domain.User
	for _, u := range m.users {
		if u.TeamName == name {
			members = append(members, u)
		}
	}
	return team, members, nil
}

func (m *mockRepo) SetUserActive(ctx context.Context, userID string, active bool) (domain.User, error) {
	u, ok := m.users[userID]
	if !ok {
		return domain.User{}, repository.ErrNotFound
	}
	u.IsActive = active
	m.users[userID] = u
	return u, nil
}

func (m *mockRepo) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	u, ok := m.users[userID]
	if !ok {
		return domain.User{}, repository.ErrNotFound
	}
	return u, nil
}

func (m *mockRepo) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	var prs []domain.PullRequest
	for prID, reviewers := range m.reviewers {
		for _, r := range reviewers {
			if r == userID {
				if pr, ok := m.prs[prID]; ok {
					prs = append(prs, pr)
				}
				break
			}
		}
	}
	return prs, nil
}

func (m *mockRepo) CreatePR(ctx context.Context, pr domain.PullRequest, status string) error {
	m.prs[pr.ID] = pr
	m.reviewers[pr.ID] = pr.Reviewers
	return nil
}

func (m *mockRepo) GetPR(ctx context.Context, prID string) (domain.PullRequest, error) {
	pr, ok := m.prs[prID]
	if !ok {
		return domain.PullRequest{}, repository.ErrNotFound
	}
	return pr, nil
}

func (m *mockRepo) GetActiveTeamMembersExcluding(ctx context.Context, teamID int, exclude []string) ([]domain.User, error) {
	var users []domain.User
	excludeSet := make(map[string]struct{})
	for _, e := range exclude {
		excludeSet[e] = struct{}{}
	}
	for _, u := range m.users {
		if u.TeamID == teamID && u.IsActive {
			if _, ex := excludeSet[u.ID]; !ex {
				users = append(users, u)
			}
		}
	}
	return users, nil
}

func (m *mockRepo) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
	return m.reviewers[prID], nil
}

func (m *mockRepo) ReplacePRReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	reviewers, ok := m.reviewers[prID]
	if !ok {
		return repository.ErrNotFound
	}
	found := false
	for i, r := range reviewers {
		if r == oldUserID {
			reviewers[i] = newUserID
			found = true
			break
		}
	}
	if !found {
		return repository.ErrNotAssigned
	}
	m.reviewers[prID] = reviewers
	return nil
}

func (m *mockRepo) MergePR(ctx context.Context, prID string) error {
	pr, ok := m.prs[prID]
	if !ok {
		return repository.ErrNotFound
	}
	pr.Status = "MERGED"
	m.prs[prID] = pr
	return nil
}

func (m *mockRepo) PRExists(ctx context.Context, prID string) (bool, error) {
	_, ok := m.prs[prID]
	return ok, nil
}

func (m *mockRepo) IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	reviewers, ok := m.reviewers[prID]
	if !ok {
		return false, nil
	}
	for _, r := range reviewers {
		if r == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) GetPRAuthor(ctx context.Context, prID string) (string, error) {
	pr, ok := m.prs[prID]
	if !ok {
		return "", repository.ErrNotFound
	}
	return pr.AuthorID, nil
}

func (m *mockRepo) HasOpenPRsAsReviewer(ctx context.Context, userID string) (bool, error) {
	for _, reviewers := range m.reviewers {
		for _, r := range reviewers {
			if r == userID {
				return true, nil
			}
		}
	}
	return false, nil
}

func (m *mockRepo) GetReviewerStats(ctx context.Context) ([]repository.ReviewerStat, error) {
	return m.stats, nil
}

func TestHealth(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handlers.Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response["status"] != "OK" {
		t.Fatalf("expected status OK, got %s", response["status"])
	}
}

func TestAddTeam_Success(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "alice", "is_active": true},
			{"user_id": "u2", "username": "bob", "is_active": true},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.AddTeam(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	team, ok := response["team"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected team in response")
	}
	if team["team_name"] != "backend" {
		t.Fatalf("expected team_name backend, got %v", team["team_name"])
	}
}

func TestAddTeam_InvalidJSON(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	handlers.AddTeam(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAddTeam_MissingTeamName(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "alice", "is_active": true},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.AddTeam(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAddTeam_TeamExists(t *testing.T) {
	repo := newMockRepo()
	repo.teams["backend"] = domain.Team{ID: 1, Name: "backend"}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "alice", "is_active": true},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/team/add", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.AddTeam(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestGetTeam_Success(t *testing.T) {
	repo := newMockRepo()
	repo.teams["backend"] = domain.Team{ID: 1, Name: "backend"}
	repo.users["u1"] = domain.User{ID: "u1", Username: "alice", TeamID: 1, TeamName: "backend", IsActive: true}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("GET", "/team/get?team_name=backend", nil)
	w := httptest.NewRecorder()
	handlers.GetTeam(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestGetTeam_MissingParam(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("GET", "/team/get", nil)
	w := httptest.NewRecorder()
	handlers.GetTeam(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestGetTeam_NotFound(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("GET", "/team/get?team_name=nonexistent", nil)
	w := httptest.NewRecorder()
	handlers.GetTeam(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestSetIsActive_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["u1"] = domain.User{ID: "u1", Username: "alice", IsActive: true}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"user_id":   "u1",
		"is_active": false,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.SetIsActive(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestSetIsActive_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"user_id":   "nonexistent",
		"is_active": false,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users/setIsActive", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.SetIsActive(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCreatePR_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["u1"] = domain.User{ID: "u1", Username: "alice", TeamID: 1, IsActive: true}
	repo.users["u2"] = domain.User{ID: "u2", Username: "bob", TeamID: 1, IsActive: true}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"pull_request_id":   "pr1",
		"pull_request_name": "feat",
		"author_id":         "u1",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.CreatePR(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestCreatePR_MissingFields(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"pull_request_id": "pr1",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.CreatePR(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"pull_request_id":   "pr1",
		"pull_request_name": "feat",
		"author_id":         "nonexistent",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pullRequest/create", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.CreatePR(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestReassign_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["u1"] = domain.User{ID: "u1", Username: "alice", TeamID: 1, IsActive: true}
	repo.users["u2"] = domain.User{ID: "u2", Username: "bob", TeamID: 1, IsActive: true}
	repo.users["u3"] = domain.User{ID: "u3", Username: "carl", TeamID: 1, IsActive: true}
	repo.prs["pr1"] = domain.PullRequest{ID: "pr1", Title: "fix", AuthorID: "u1", Status: "OPEN"}
	repo.reviewers["pr1"] = []string{"u2"}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"pull_request_id": "pr1",
		"old_user_id":     "u2",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.Reassign(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestReassign_PRNotFound(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"pull_request_id": "nonexistent",
		"old_user_id":     "u1",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pullRequest/reassign", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.Reassign(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestMerge_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["u1"] = domain.User{ID: "u1", Username: "alice", TeamID: 1, IsActive: true}
	repo.prs["pr1"] = domain.PullRequest{ID: "pr1", Title: "merge", AuthorID: "u1", Status: "OPEN"}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"pull_request_id": "pr1",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.Merge(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestMerge_PRNotFound(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	payload := map[string]interface{}{
		"pull_request_id": "nonexistent",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/pullRequest/merge", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handlers.Merge(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestGetUserReviews_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["u1"] = domain.User{ID: "u1", Username: "alice", IsActive: true}
	repo.prs["pr1"] = domain.PullRequest{ID: "pr1", Title: "feat", AuthorID: "u2", Status: "OPEN"}
	repo.reviewers["pr1"] = []string{"u1"}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("GET", "/users/getReview?user_id=u1", nil)
	w := httptest.NewRecorder()
	handlers.GetUserReviews(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestGetUserReviews_MissingParam(t *testing.T) {
	repo := newMockRepo()
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("GET", "/users/getReview", nil)
	w := httptest.NewRecorder()
	handlers.GetUserReviews(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestGetStats_Success(t *testing.T) {
	repo := newMockRepo()
	repo.stats = []repository.ReviewerStat{
		{UserID: "u1", Username: "alice", Count: 5},
		{UserID: "u2", Username: "bob", Count: 3},
	}
	ucase := uc.NewPRUsecase(repo)
	logger := infra.NewStdLogger()
	handlers := NewHandlers(ucase, repo, logger)

	req := httptest.NewRequest("GET", "/statistics/reviewers", nil)
	w := httptest.NewRecorder()
	handlers.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	stats, ok := response["statistics"].([]interface{})
	if !ok {
		t.Fatalf("expected statistics in response")
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
}
