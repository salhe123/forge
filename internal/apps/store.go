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
var ErrNoImage = errors.New("app image is required")

const appCols = `id::text, name, repo_url, image, status, last_error, created_at, updated_at`

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func scanApp(row interface{ Scan(dest ...any) error }) (App, error) {
	var app App
	err := row.Scan(&app.ID, &app.Name, &app.RepoURL, &app.Image, &app.Status, &app.LastError, &app.CreatedAt, &app.UpdatedAt)
	return app, err
}

func (s *Store) Create(ctx context.Context, in CreateAppInput) (App, error) {
	q := `
INSERT INTO apps (name, repo_url, image)
VALUES ($1, $2, $3)
RETURNING ` + appCols
	app, err := scanApp(s.pool.QueryRow(ctx, q, in.Name, in.RepoURL, in.Image))
	if isUniqueViolation(err) {
		return App{}, ErrConflict
	}
	if err != nil {
		return App{}, fmt.Errorf("create app: %w", err)
	}
	return app, nil
}

func (s *Store) List(ctx context.Context) ([]App, error) {
	q := `SELECT ` + appCols + ` FROM apps ORDER BY created_at DESC`
	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list apps: %w", err)
	}
	defer rows.Close()

	out := make([]App, 0)
	for rows.Next() {
		app, err := scanApp(rows)
		if err != nil {
			return nil, fmt.Errorf("scan app: %w", err)
		}
		out = append(out, app)
	}
	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, id string) (App, error) {
	q := `SELECT ` + appCols + ` FROM apps WHERE id = $1`
	app, err := scanApp(s.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return App{}, ErrNotFound
	}
	if err != nil {
		return App{}, fmt.Errorf("get app: %w", err)
	}
	return app, nil
}

func (s *Store) UpdateStatus(ctx context.Context, id, status, lastError string) (App, error) {
	q := `
UPDATE apps
SET status = $2, last_error = $3, updated_at = now()
WHERE id = $1
RETURNING ` + appCols
	app, err := scanApp(s.pool.QueryRow(ctx, q, id, status, lastError))
	if errors.Is(err, pgx.ErrNoRows) {
		return App{}, ErrNotFound
	}
	if err != nil {
		return App{}, fmt.Errorf("update status: %w", err)
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
