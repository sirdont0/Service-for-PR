# PR Assign Service (Avito-style)

## Quick start (recommended)
1. Install Docker & docker-compose.
2. Run:
   ```bash
   docker-compose up --build
   ```
3. Wait till migrate finishes and app starts.
4. Health:
   ```bash
   curl http://localhost:8080/health
   # {"status":"OK"}
   ```

## Endpoints (examples)

* Create team:
  ```
  POST /team/add
  {
    "team_name":"backend",
    "members":[{"user_id":"u1","username":"alice","is_active":true},{"user_id":"u2","username":"bob","is_active":true}]
  }
  ```
* Create PR:
  ```
  POST /pullRequest/create
  {
    "pull_request_id":"pr1",
    "pull_request_name":"feat",
    "author_id":"u1"
  }
  ```
* Reassign:
  ```
  POST /pullRequest/reassign
  {
    "pull_request_id":"pr1",
    "old_user_id":"u2"
  }
  ```
* Merge:
  ```
  POST /pullRequest/merge
  {"pull_request_id":"pr1"}
  ```

## Tests

* Unit tests (business logic): `make test`

## Notes about design

* Business rules reside in `internal/usecase`.
* Repositories contain only SQL.
* PR statuses stored in separate table (normalization).
* Transactions and `FOR UPDATE` used to prevent race conditions for PR modifications.
