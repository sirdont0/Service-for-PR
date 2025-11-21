package http

import (
	"encoding/json"
	"net/http"

	"github.com/you/pr-assign-avito/internal/domain"
	"github.com/you/pr-assign-avito/internal/infra"
	"github.com/you/pr-assign-avito/internal/repository"
	uc "github.com/you/pr-assign-avito/internal/usecase"
)

const (
	codeTeamExists  = "TEAM_EXISTS"
	codePRExists    = "PR_EXISTS"
	codePRMerged    = "PR_MERGED"
	codeNotAssigned = "NOT_ASSIGNED"
	codeNoCandidate = "NO_CANDIDATE"
	codeNotFound    = "NOT_FOUND"
)

type Handlers struct {
	UC   *uc.PRUsecase
	Repo repository.Repo
	Log  infra.Logger
}

type apiTeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type apiTeam struct {
	TeamName string          `json:"team_name"`
	Members  []apiTeamMember `json:"members"`
}

type apiUser struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type apiPullRequestShort struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

func NewHandlers(uc *uc.PRUsecase, repo repository.Repo, log infra.Logger) *Handlers {
	return &Handlers{UC: uc, Repo: repo, Log: log}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func errorResp(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}

func badRequest(w http.ResponseWriter, msg string) {
	errorResp(w, http.StatusBadRequest, codeNotFound, msg)
}

func notFound(w http.ResponseWriter, msg string) {
	errorResp(w, http.StatusNotFound, codeNotFound, msg)
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "OK"})
}

func (h *Handlers) AddTeam(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		TeamName string `json:"team_name"`
		Members  []struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			IsActive bool   `json:"is_active"`
		} `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if payload.TeamName == "" {
		badRequest(w, "team_name required")
		return
	}
	users := make([]domain.User, 0, len(payload.Members))
	for _, m := range payload.Members {
		if m.UserID == "" || m.Username == "" {
			badRequest(w, "member user_id and username required")
			return
		}
		users = append(users, domain.User{ID: m.UserID, Username: m.Username, IsActive: m.IsActive})
	}
	if err := h.Repo.CreateTeamWithMembers(r.Context(), payload.TeamName, users); err != nil {
		if err == repository.ErrTeamExists {
			errorResp(w, http.StatusBadRequest, codeTeamExists, payload.TeamName+" already exists")
			return
		}
		errorResp(w, http.StatusInternalServerError, codeNotFound, "internal server error")
		return
	}
	team, members, err := h.Repo.GetTeamByName(r.Context(), payload.TeamName)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, codeNotFound, "internal server error")
		return
	}
	apiTeamResp := buildAPITeam(team, members)
	resp := struct {
		Team apiTeam `json:"team"`
	}{Team: apiTeamResp}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("team_name")
	if q == "" {
		badRequest(w, "team_name required")
		return
	}
	team, users, err := h.Repo.GetTeamByName(r.Context(), q)
	if err != nil {
		notFound(w, "team not found")
		return
	}
	writeJSON(w, http.StatusOK, buildAPITeam(team, users))
}

func (h *Handlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if payload.UserID == "" {
		badRequest(w, "user_id required")
		return
	}
	user, err := h.Repo.SetUserActive(r.Context(), payload.UserID, payload.IsActive)
	if err != nil {
		notFound(w, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": buildAPIUser(user)})
}

func (h *Handlers) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("user_id")
	if uid == "" {
		badRequest(w, "user_id required")
		return
	}
	if _, err := h.Repo.GetUserByID(r.Context(), uid); err != nil {
		notFound(w, "user not found")
		return
	}
	prs, err := h.UC.Repo.GetUserReviews(r.Context(), uid)
	if err != nil {
		notFound(w, "user not found")
		return
	}
	short := make([]apiPullRequestShort, 0, len(prs))
	for _, pr := range prs {
		short = append(short, apiPullRequestShort{
			ID:       pr.ID,
			Name:     pr.Title,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user_id": uid, "pull_requests": short})
}

func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if payload.PullRequestID == "" || payload.PullRequestName == "" || payload.AuthorID == "" {
		badRequest(w, "pull_request_id, pull_request_name and author_id required")
		return
	}
	pr := domain.PullRequest{
		ID:       payload.PullRequestID,
		Title:    payload.PullRequestName,
		AuthorID: payload.AuthorID,
	}
	created, err := h.UC.CreatePR(r.Context(), pr)
	if err != nil {
		switch err {
		case uc.ErrPRExists:
			errorResp(w, http.StatusConflict, codePRExists, "PR id already exists")
		case uc.ErrNotFound:
			notFound(w, "author or team not found")
		default:
			errorResp(w, http.StatusInternalServerError, codeNotFound, "internal server error")
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"pr": created})
}

func (h *Handlers) Reassign(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if payload.PullRequestID == "" || payload.OldUserID == "" {
		badRequest(w, "pull_request_id and old_user_id required")
		return
	}
	newID, err := h.UC.ReassignReviewer(r.Context(), payload.PullRequestID, payload.OldUserID)
	if err != nil {
		switch err {
		case uc.ErrNotFound:
			notFound(w, "PR or user not found")
		case uc.ErrPRMerged:
			errorResp(w, http.StatusConflict, codePRMerged, "cannot reassign on merged PR")
		case uc.ErrNotAssigned:
			errorResp(w, http.StatusConflict, codeNotAssigned, "reviewer is not assigned to this PR")
		case uc.ErrNoCandidate:
			errorResp(w, http.StatusConflict, codeNoCandidate, "no active replacement candidate in team")
		default:
			errorResp(w, http.StatusInternalServerError, codeNotFound, "internal server error")
		}
		return
	}
	pr, err := h.UC.Repo.GetPR(r.Context(), payload.PullRequestID)
	if err != nil {
		errorResp(w, http.StatusInternalServerError, codeNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"pr": pr, "replaced_by": newID})
}

func (h *Handlers) Merge(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PullRequestID string `json:"pull_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		badRequest(w, "invalid json")
		return
	}
	if payload.PullRequestID == "" {
		badRequest(w, "pull_request_id required")
		return
	}
	pr, err := h.UC.MergePR(r.Context(), payload.PullRequestID)
	if err != nil {
		if err == uc.ErrNotFound {
			notFound(w, "PR not found")
			return
		}
		errorResp(w, http.StatusInternalServerError, codeNotFound, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"pr": pr})
}

func buildAPITeam(team domain.Team, members []domain.User) apiTeam {
	resp := apiTeam{
		TeamName: team.Name,
		Members:  make([]apiTeamMember, 0, len(members)),
	}
	for _, m := range members {
		resp.Members = append(resp.Members, apiTeamMember{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	return resp
}

func buildAPIUser(u domain.User) apiUser {
	return apiUser{
		UserID:   u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}
