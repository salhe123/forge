package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/salhe123/forge/internal/apps"
	"github.com/salhe123/forge/internal/deploy"
)

type Server struct {
	router *chi.Mux
	pool   *pgxpool.Pool
	apps   *apps.Store
	deploy *deploy.Service
}

func New(pool *pgxpool.Pool, store *apps.Store, deploys *deploy.Service) *Server {
	s := &Server{
		router: chi.NewRouter(),
		pool:   pool,
		apps:   store,
		deploy: deploys,
	}
	s.routes()
	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) routes() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Get("/health", s.health)
	s.router.Get("/ready", s.ready)
	s.router.Handle("/metrics", promhttp.Handler())

	s.router.Route("/v1/apps", func(r chi.Router) {
		r.Get("/", s.listApps)
		r.Post("/", s.createApp)
		r.Get("/{id}", s.getApp)
		r.Post("/{id}/deploy", s.deployApp)
	})
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.pool.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "error": "database unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) createApp(w http.ResponseWriter, r *http.Request) {
	var in apps.CreateAppInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	app, err := s.apps.Create(r.Context(), in)
	if errors.Is(err, apps.ErrConflict) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "app name already exists"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not create app"})
		return
	}
	writeJSON(w, http.StatusCreated, app)
}

func (s *Server) listApps(w http.ResponseWriter, r *http.Request) {
	list, err := s.apps.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not list apps"})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) getApp(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	app, err := s.apps.Get(r.Context(), id)
	if errors.Is(err, apps.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "app not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not get app"})
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) deployApp(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	app, err := s.deploy.Start(r.Context(), id)
	if errors.Is(err, apps.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "app not found"})
		return
	}
	if errors.Is(err, apps.ErrNoImage) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "image is required to deploy"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "could not start deploy"})
		return
	}
	writeJSON(w, http.StatusAccepted, app)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
