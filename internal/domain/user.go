package domain

type User struct {
    ID       string `json:"user_id"`
    Username string `json:"username"`
    TeamID   int    `json:"team_id"`
    IsActive bool   `json:"is_active"`
}
