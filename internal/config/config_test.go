package config

import (
	"os"
	"testing"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://forge:forge@localhost:5432/forge?sslmode=disable")
	os.Unsetenv("HTTP_ADDR")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %s", cfg.HTTPAddr)
	}
}
