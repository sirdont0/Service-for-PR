package domain

type Team struct {
    ID   int    `json:"-"`
    Name string `json:"team_name"`
}
