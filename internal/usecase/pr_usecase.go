package usecase

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/you/pr-assign-avito/internal/domain"
	"github.com/you/pr-assign-avito/internal/repository"
)

var (
	ErrPRExists    = errors.New("pr exists")
	ErrNotFound    = errors.New("not found")
	ErrPRMerged    = errors.New("pr merged")
	ErrNotAssigned = errors.New("not assigned")
	ErrNoCandidate = errors.New("no candidate")
)

type PRUsecase struct {
	Repo repository.Repo
	rand *rand.Rand
}

func NewPRUsecase(r repository.Repo) *PRUsecase {
	return &PRUsecase{
		Repo: r,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (u *PRUsecase) CreatePR(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	exists, err := u.Repo.PRExists(ctx, pr.ID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	if exists {
		return domain.PullRequest{}, ErrPRExists
	}

	author, err := u.Repo.GetUserByID(ctx, pr.AuthorID)
	if err != nil {
		return domain.PullRequest{}, ErrNotFound
	}

	candidates, err := u.Repo.GetActiveTeamMembersExcluding(ctx, author.TeamID, []string{author.ID})
	if err != nil {
		return domain.PullRequest{}, err
	}

	ids := make([]string, 0, len(candidates))
	for _, c := range candidates {
		ids = append(ids, c.ID)
	}
	u.shuffle(ids)
	chosen := pickUpTo(ids, 2)

	pr.Reviewers = chosen
	pr.Status = "OPEN"
	pr.CreatedAt = time.Now().UTC()

	if err := u.Repo.CreatePR(ctx, pr, "OPEN"); err != nil {
		return domain.PullRequest{}, err
	}
	return pr, nil
}

func (u *PRUsecase) ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, error) {
	assigned, err := u.Repo.IsReviewerAssigned(ctx, prID, oldUserID)
	if err != nil {
		return "", err
	}
	if !assigned {
		return "", ErrNotAssigned
	}

	oldUser, err := u.Repo.GetUserByID(ctx, oldUserID)
	if err != nil {
		return "", ErrNotFound
	}

	current, err := u.Repo.GetPRReviewers(ctx, prID)
	if err != nil {
		return "", err
	}

	author, err := u.Repo.GetPRAuthor(ctx, prID)
	if err != nil {
		return "", err
	}

	excludeSet := map[string]struct{}{}
	excludeSet[author] = struct{}{}
	for _, c := range current {
		excludeSet[c] = struct{}{}
	}
	excludeSet[oldUserID] = struct{}{}

	var excludeList []string
	for k := range excludeSet {
		excludeList = append(excludeList, k)
	}

	cands, err := u.Repo.GetActiveTeamMembersExcluding(ctx, oldUser.TeamID, excludeList)
	if err != nil {
		return "", err
	}
	if len(cands) == 0 {
		return "", ErrNoCandidate
	}
	ids := []string{}
	for _, c := range cands {
		ids = append(ids, c.ID)
	}
	u.shuffle(ids)
	newID := ids[0]

	if err := u.Repo.ReplacePRReviewer(ctx, prID, oldUserID, newID); err != nil {
		switch err {
		case repository.ErrNotFound:
			return "", ErrNotFound
		case repository.ErrPRMerged:
			return "", ErrPRMerged
		case repository.ErrNotAssigned:
			return "", ErrNotAssigned
		default:
			return "", err
		}
	}
	return newID, nil
}

func (u *PRUsecase) MergePR(ctx context.Context, prID string) (domain.PullRequest, error) {
	if err := u.Repo.MergePR(ctx, prID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.PullRequest{}, ErrNotFound
		}
		return domain.PullRequest{}, err
	}
	pr, err := u.Repo.GetPR(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	return pr, nil
}

func (u *PRUsecase) shuffle(ids []string) {
	for i := range ids {
		j := u.rand.Intn(i + 1)
		ids[i], ids[j] = ids[j], ids[i]
	}
}

func pickUpTo(ids []string, n int) []string {
	if n >= len(ids) {
		return ids
	}
	return ids[:n]
}
