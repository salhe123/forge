package apps

import "time"

type App struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
	Image     string    `json:"image"`
	Status    string    `json:"status"`
	LastError string    `json:"last_error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAppInput struct {
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
	Image   string `json:"image"`
}
