package apps

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("app not found")
var ErrConflict = errors.New("app already exists")

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Create(ctx context.Context, in CreateAppInput) (App, error) {
	const q = `
INSERT INTO apps (name, repo_url, image)
VALUES ($1, $2, $3)
RETURNING id::text, name, repo_url, image, status, created_at, updated_at
`
	var app App
	err := s.pool.QueryRow(ctx, q, in.Name, in.RepoURL, in.Image).Scan(
		&app.ID, &app.Name, &app.RepoURL, &app.Image, &app.Status, &app.CreatedAt, &app.UpdatedAt,
	)
	if isUniqueViolation(err) {
		return App{}, ErrConflict
	}
	if err != nil {
		return App{}, fmt.Errorf("create app: %w", err)
	}
	return app, nil
}

func (s *Store) List(ctx context.Context) ([]App, error) {
	const q = `
SELECT id::text, name, repo_url, image, status, created_at, updated_at
FROM apps
ORDER BY created_at DESC
`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list apps: %w", err)
	}
	defer rows.Close()

	out := make([]App, 0)
	for rows.Next() {
		var app App
		if err := rows.Scan(&app.ID, &app.Name, &app.RepoURL, &app.Image, &app.Status, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan app: %w", err)
		}
		out = append(out, app)
	}
	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, id string) (App, error) {
	const q = `
SELECT id::text, name, repo_url, image, status, created_at, updated_at
FROM apps
WHERE id = $1
`
	var app App
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&app.ID, &app.Name, &app.RepoURL, &app.Image, &app.Status, &app.CreatedAt, &app.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return App{}, ErrNotFound
	}
	if err != nil {
		return App{}, fmt.Errorf("get app: %w", err)
	}
	return app, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}
