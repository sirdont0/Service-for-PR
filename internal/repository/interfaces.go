package repository

import (
	"context"
	"errors"

	"github.com/you/pr-assign-avito/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Repo interface {
	CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.User) error
	GetTeamByName(ctx context.Context, name string) (domain.Team, []domain.User, error)
	SetUserActive(ctx context.Context, userID string, active bool) (domain.User, error)
	GetUserByID(ctx context.Context, userID string) (domain.User, error)
	GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequest, error)

	CreatePR(ctx context.Context, pr domain.PullRequest, status string) error
	GetPR(ctx context.Context, prID string) (domain.PullRequest, error)
	GetActiveTeamMembersExcluding(ctx context.Context, teamID int, exclude []string) ([]domain.User, error)
	GetPRReviewers(ctx context.Context, prID string) ([]string, error)
	ReplacePRReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	DeletePRReviewer(ctx context.Context, prID, userID string) error
	AddPRReviewer(ctx context.Context, prID, userID string) error
	MergePR(ctx context.Context, prID string) error
	PRExists(ctx context.Context, prID string) (bool, error)
	IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error)
	GetPRAuthor(ctx context.Context, prID string) (string, error)
	LockPRForUpdate(ctx context.Context, prID string) (string, error)
}
