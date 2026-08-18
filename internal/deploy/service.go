package deploy

import (
	"context"
	"log/slog"

	"github.com/salhe123/forge/internal/apps"
)

type Service struct {
	store  *apps.Store
	runner Runner
}

func NewService(store *apps.Store, runner Runner) *Service {
	return &Service{store: store, runner: runner}
}

func (s *Service) Start(ctx context.Context, id string) (apps.App, error) {
	app, err := s.store.Get(ctx, id)
	if err != nil {
		return apps.App{}, err
	}
	if app.Image == "" {
		return apps.App{}, apps.ErrNoImage
	}

	app, err = s.store.UpdateStatus(ctx, id, apps.StatusDeploying, "")
	if err != nil {
		return apps.App{}, err
	}

	go s.run(id, app.Name, app.Image)
	return app, nil
}

func (s *Service) run(id, name, image string) {
	ctx := context.Background()
	if err := s.runner.Run(ctx, name, image); err != nil {
		slog.Error("deploy failed", "app_id", id, "err", err)
		if _, uerr := s.store.UpdateStatus(ctx, id, apps.StatusFailed, err.Error()); uerr != nil {
			slog.Error("status update failed", "app_id", id, "err", uerr)
		}
		return
	}
	if _, err := s.store.UpdateStatus(ctx, id, apps.StatusRunning, ""); err != nil {
		slog.Error("status update failed", "app_id", id, "err", err)
	}
}
