исправить по линтерам

tomchukd@MacBook-Pro-Dmitrij avito % golangci-lint run --config=.golangci.yml

internal/repository/pg/pg_repo.go:31:19: Error return value of `tx.Rollback` is not checked (errcheck)
	defer tx.Rollback(ctx)
	                 ^
internal/repository/pg/pg_repo.go:128:19: Error return value of `tx.Rollback` is not checked (errcheck)
	defer tx.Rollback(ctx)
	                 ^
internal/repository/pg/pg_repo.go:182:13: Error return value of `rows.Scan` is not checked (errcheck)
			rows.Scan(&rid)
			         ^
internal/repository/pg/pg_repo.go:245:12: Error return value of `rows.Scan` is not checked (errcheck)
		rows.Scan(&r)
		         ^
internal/repository/pg/pg_repo.go:302:19: Error return value of `tx.Rollback` is not checked (errcheck)
	defer tx.Rollback(ctx)
	                 ^
internal/usecase/pr_usecase_test.go:209:28: Error return value of `repo.CreateTeamWithMembers` is not checked (errcheck)
	repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
	                          ^
internal/usecase/pr_usecase_test.go:231:28: Error return value of `repo.CreateTeamWithMembers` is not checked (errcheck)
	repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
	                          ^
internal/usecase/pr_usecase_test.go:251:28: Error return value of `repo.CreateTeamWithMembers` is not checked (errcheck)
	repo.CreateTeamWithMembers(ctx, "backend", []domain.User{
	                          ^
cmd/server/main.go:36:3: exitAfterDefer: log.Fatalf will exit, and `defer cancel()` will not run (gocritic)
		log.Fatalf("db connect: %v", err)
		^
internal/usecase/pr_usecase.go:30:9: G404: Use of weak random number generator (math/rand or math/rand/v2 instead of crypto/rand) (gosec)
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		      ^
internal/domain/pr.go:1:1: package-comments: should have a package comment (revive)
package domain
^
internal/domain/pr.go:5:6: exported: exported type PullRequest should have comment or be unexported (revive)
type PullRequest struct {
     ^
internal/domain/team.go:3:6: exported: exported type Team should have comment or be unexported (revive)
type Team struct {
     ^
internal/domain/user.go:3:6: exported: exported type User should have comment or be unexported (revive)
type User struct {
     ^
internal/infra/logger.go:1:1: package-comments: should have a package comment (revive)
package infra
^
internal/infra/logger.go:5:6: exported: exported type Logger should have comment or be unexported (revive)
type Logger interface {
     ^
internal/infra/logger.go:12:1: exported: exported function NewStdLogger should have comment or be unexported (revive)
func NewStdLogger() Logger { return &stdLogger{} }
^
internal/repository/interfaces.go:1:1: package-comments: should have a package comment (revive)
package repository
^
internal/repository/interfaces.go:11:2: exported: exported var ErrNotFound should have comment or be unexported (revive)
	ErrNotFound    = errors.New("not found")
	^
internal/repository/interfaces.go:17:6: exported: exported type Repo should have comment or be unexported (revive)
type Repo interface {
     ^
internal/repository/pg/pg_repo.go:18:6: exported: exported type PGRepo should have comment or be unexported (revive)
type PGRepo struct {
     ^
internal/repository/pg/pg_repo.go:22:1: exported: exported function NewPGRepo should have comment or be unexported (revive)
func NewPGRepo(pool *pgxpool.Pool) *PGRepo {
^
internal/repository/pg/pg_repo.go:26:1: exported: exported method PGRepo.CreateTeamWithMembers should have comment or be unexported (revive)
func (p *PGRepo) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.User) error {
^
internal/repository/pg/pg_repo.go:57:1: exported: exported method PGRepo.GetTeamByName should have comment or be unexported (revive)
func (p *PGRepo) GetTeamByName(ctx context.Context, name string) (domain.Team, []domain.User, error) {
^
internal/repository/pg/pg_repo.go:79:1: exported: exported method PGRepo.SetUserActive should have comment or be unexported (revive)
func (p *PGRepo) SetUserActive(ctx context.Context, userID string, active bool) (domain.User, error) {
^
internal/repository/pg/pg_repo.go:100:1: exported: exported method PGRepo.GetUserByID should have comment or be unexported (revive)
func (p *PGRepo) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
^
internal/repository/pg/pg_repo.go:114:1: exported: exported method PGRepo.PRExists should have comment or be unexported (revive)
func (p *PGRepo) PRExists(ctx context.Context, prID string) (bool, error) {
^
internal/repository/pg/pg_repo.go:123:1: exported: exported method PGRepo.CreatePR should have comment or be unexported (revive)
func (p *PGRepo) CreatePR(ctx context.Context, pr domain.PullRequest, status string) error {
^
internal/repository/pg/pg_repo.go:157:1: exported: exported method PGRepo.GetPR should have comment or be unexported (revive)
func (p *PGRepo) GetPR(ctx context.Context, prID string) (domain.PullRequest, error) {
^
internal/repository/pg/pg_repo.go:209:1: exported: exported method PGRepo.GetActiveTeamMembersExcluding should have comment or be unexported (revive)
func (p *PGRepo) GetActiveTeamMembersExcluding(ctx context.Context, teamID int, exclude []string) ([]domain.User, error) {
^
internal/repository/pg/pg_repo.go:236:1: exported: exported method PGRepo.GetPRReviewers should have comment or be unexported (revive)
func (p *PGRepo) GetPRReviewers(ctx context.Context, prID string) ([]string, error) {
^
internal/repository/pg/pg_repo.go:251:1: exported: exported method PGRepo.IsReviewerAssigned should have comment or be unexported (revive)
func (p *PGRepo) IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
^
internal/repository/pg/pg_repo.go:260:1: exported: exported method PGRepo.GetPRAuthor should have comment or be unexported (revive)
func (p *PGRepo) GetPRAuthor(ctx context.Context, prID string) (string, error) {
^
internal/repository/pg/pg_repo.go:269:1: exported: exported method PGRepo.GetUserReviews should have comment or be unexported (revive)
func (p *PGRepo) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequest, error) {
^
internal/repository/pg/pg_repo.go:297:1: exported: exported method PGRepo.ReplacePRReviewer should have comment or be unexported (revive)
func (p *PGRepo) ReplacePRReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
^
internal/repository/pg/pg_repo.go:367:1: exported: exported method PGRepo.MergePR should have comment or be unexported (revive)
func (p *PGRepo) MergePR(ctx context.Context, prID string) error {
^
internal/repository/pg/pg_repo.go:405:1: exported: exported method PGRepo.HasOpenPRsAsReviewer should have comment or be unexported (revive)
func (p *PGRepo) HasOpenPRsAsReviewer(ctx context.Context, userID string) (bool, error) {
^
internal/transport/http/handlers.go:23:6: exported: exported type Handlers should have comment or be unexported (revive)
type Handlers struct {
     ^
internal/transport/http/handlers.go:54:1: exported: exported function NewHandlers should have comment or be unexported (revive)
func NewHandlers(uc *uc.PRUsecase, repo repository.Repo, log infra.Logger) *Handlers {
^
internal/transport/http/handlers.go:81:1: exported: exported method Handlers.Health should have comment or be unexported (revive)
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/handlers.go:85:1: exported: exported method Handlers.AddTeam should have comment or be unexported (revive)
func (h *Handlers) AddTeam(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/handlers.go:133:1: exported: exported method Handlers.GetTeam should have comment or be unexported (revive)
func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/handlers.go:152:1: exported: exported method Handlers.SetIsActive should have comment or be unexported (revive)
func (h *Handlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/handlers.go:179:1: exported: exported method Handlers.GetUserReviews should have comment or be unexported (revive)
func (h *Handlers) GetUserReviews(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/handlers.go:212:1: exported: exported method Handlers.CreatePR should have comment or be unexported (revive)
func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/handlers.go:248:1: exported: exported method Handlers.Reassign should have comment or be unexported (revive)
func (h *Handlers) Reassign(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/handlers.go:288:1: exported: exported method Handlers.Merge should have comment or be unexported (revive)
func (h *Handlers) Merge(w http.ResponseWriter, r *http.Request) {
^
internal/transport/http/router.go:9:1: exported: exported function NewRouter should have comment or be unexported (revive)
func NewRouter(h *Handlers) http.Handler {
^
internal/usecase/pr_usecase.go:14:2: exported: exported var ErrPRExists should have comment or be unexported (revive)
	ErrPRExists    = errors.New("pr exists")
	^
internal/usecase/pr_usecase.go:22:6: exported: exported type PRUsecase should have comment or be unexported (revive)
type PRUsecase struct {
     ^
internal/usecase/pr_usecase.go:27:1: exported: exported function NewPRUsecase should have comment or be unexported (revive)
func NewPRUsecase(r repository.Repo) *PRUsecase {
^
internal/usecase/pr_usecase.go:34:1: exported: exported method PRUsecase.CreatePR should have comment or be unexported (revive)
func (u *PRUsecase) CreatePR(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
^
internal/usecase/pr_usecase.go:70:1: exported: exported method PRUsecase.ReassignReviewer should have comment or be unexported (revive)
func (u *PRUsecase) ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, error) {
^
internal/usecase/pr_usecase.go:150:1: exported: exported method PRUsecase.MergePR should have comment or be unexported (revive)
func (u *PRUsecase) MergePR(ctx context.Context, prID string) (domain.PullRequest, error) {
^
internal/usecase/pr_usecase_test.go:31:41: unused-parameter: parameter 'ctx' seems to be unused, consider removing or renaming it as _ (revive)
func (m *memRepo) CreateTeamWithMembers(ctx context.Context, teamName string, members []domain.User) error {
                                        ^
internal/usecase/pr_usecase_test.go:44:33: unused-parameter: parameter 'ctx' seems to be unused, consider removing or renaming it as _ (revive)
func (m *memRepo) GetTeamByName(ctx context.Context, name string) (domain.Team, []domain.User, error) {
                                ^
internal/usecase/pr_usecase_test.go:57:33: unused-parameter: parameter 'ctx' seems to be unused, consider removing or renaming it as _ (revive)
func (m *memRepo) SetUserActive(ctx context.Context, userID string, active bool) (domain.User, error) {
                                ^
57 issues:
* errcheck: 8
* gocritic: 1
* gosec: 1
* revive: 47
