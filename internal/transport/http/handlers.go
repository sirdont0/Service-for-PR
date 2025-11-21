package http

import (
    "encoding/json"
    "net/http"

    uc "github.com/you/pr-assign-avito/internal/usecase"
    "github.com/you/pr-assign-avito/internal/domain"
    "github.com/you/pr-assign-avito/internal/infra"
    "github.com/you/pr-assign-avito/internal/repository"
)

type Handlers struct {
    UC *uc.PRUsecase
    Repo repository.Repo
    Log  infra.Logger
}

func NewHandlers(uc *uc.PRUsecase, repo repository.Repo, log infra.Logger) *Handlers {
    return &Handlers{UC:uc, Repo:repo, Log:log}
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(v)
}

func errorResp(w http.ResponseWriter, status int, code, msg string) {
    writeJSON(w, status, map[string]interface{}{"error": map[string]string{"code": code, "message": msg}})
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status":"OK"})
}

func (h *Handlers) AddTeam(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        TeamName string `json:"team_name"`
        Members  []struct {
            UserID string `json:"user_id"`
            Username string `json:"username"`
            IsActive bool `json:"is_active"`
        } `json:"members"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        errorResp(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
        return
    }
    users := []domain.User{}
    for _, m := range payload.Members {
        users = append(users, domain.User{ID:m.UserID, Username:m.Username, IsActive:m.IsActive})
    }
    if err := h.Repo.CreateTeamWithMembers(r.Context(), payload.TeamName, users); err != nil {
        errorResp(w, http.StatusBadRequest, "TEAM_EXISTS", err.Error())
        return
    }
    writeJSON(w, http.StatusCreated, map[string]interface{}{"team": payload.TeamName})
}

func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query().Get("team_name")
    if q=="" { errorResp(w, http.StatusBadRequest,"BAD_REQUEST","team_name required"); return }
    team, users, err := h.Repo.GetTeamByName(r.Context(), q)
    if err != nil {
        errorResp(w, http.StatusNotFound, "NOT_FOUND", err.Error()); return
    }
    resp := map[string]interface{}{"team_name": team.Name, "members": users}
    writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        UserID string `json:"user_id"`
        IsActive bool `json:"is_active"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        errorResp(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json"); return
    }
    user, err := h.Repo.SetUserActive(r.Context(), payload.UserID, payload.IsActive)
    if err != nil {
        errorResp(w, http.StatusNotFound, "NOT_FOUND", err.Error()); return
    }
    writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (h *Handlers) GetUserReviews(w http.ResponseWriter, r *http.Request) {
    uid := r.URL.Query().Get("user_id")
    if uid=="" { errorResp(w, http.StatusBadRequest,"BAD_REQUEST","user_id required"); return }
    prs, err := h.UC.Repo.GetUserReviews(r.Context(), uid)
    if err != nil {
        errorResp(w, http.StatusNotFound, "NOT_FOUND", err.Error()); return
    }
    writeJSON(w, http.StatusOK, map[string]interface{}{"user_id": uid, "pull_requests": prs})
}

func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        PullRequestID   string `json:"pull_request_id"`
        PullRequestName string `json:"pull_request_name"`
        AuthorID        string `json:"author_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        errorResp(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json"); return
    }
    pr := domain.PullRequest{
        ID: payload.PullRequestID,
        Title: payload.PullRequestName,
        AuthorID: payload.AuthorID,
    }
    created, err := h.UC.CreatePR(r.Context(), pr)
    if err != nil {
        switch err {
        case uc.ErrPRExists:
            errorResp(w, http.StatusConflict, "PR_EXISTS", err.Error())
        case uc.ErrNotFound:
            errorResp(w, http.StatusNotFound, "NOT_FOUND", err.Error())
        default:
            errorResp(w, http.StatusInternalServerError, "INTERNAL", err.Error())
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
        errorResp(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json"); return
    }
    newID, err := h.UC.ReassignReviewer(r.Context(), payload.PullRequestID, payload.OldUserID)
    if err != nil {
        switch err {
        case uc.ErrNotFound:
            errorResp(w, http.StatusNotFound,"NOT_FOUND",err.Error())
        case uc.ErrPRMerged:
            errorResp(w, http.StatusConflict,"PR_MERGED",err.Error())
        case uc.ErrNotAssigned:
            errorResp(w, http.StatusConflict,"NOT_ASSIGNED",err.Error())
        case uc.ErrNoCandidate:
            errorResp(w, http.StatusConflict,"NO_CANDIDATE",err.Error())
        default:
            errorResp(w, http.StatusInternalServerError,"INTERNAL",err.Error())
        }
        return
    }
    pr, _ := h.UC.Repo.GetPR(r.Context(), payload.PullRequestID)
    writeJSON(w, http.StatusOK, map[string]interface{}{"pr": pr, "replaced_by": newID})
}

func (h *Handlers) Merge(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        PullRequestID string `json:"pull_request_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        errorResp(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json"); return
    }
    pr, err := h.UC.MergePR(r.Context(), payload.PullRequestID)
    if err != nil {
        if err == uc.ErrNotFound {
            errorResp(w, http.StatusNotFound, "NOT_FOUND", err.Error()); return
        }
        errorResp(w, http.StatusInternalServerError, "INTERNAL", err.Error()); return
    }
    writeJSON(w, http.StatusOK, map[string]interface{}{"pr": pr})
}
