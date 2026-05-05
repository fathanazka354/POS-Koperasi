package config

import (
	"os"
	"testing"
	"time"
)

func TestParseDurationEnv(t *testing.T) {
	t.Setenv("TEST_PARSE_DUR", "45s")
	if d := parseDurationEnv("TEST_PARSE_DUR", time.Second); d != 45*time.Second {
		t.Fatalf("got %v want 45s", d)
	}
	if d := parseDurationEnv("TEST_PARSE_DUR_MISSING", 2*time.Minute); d != 2*time.Minute {
		t.Fatalf("default got %v", d)
	}
	t.Setenv("TEST_PARSE_DUR_ZERO", "0")
	if d := parseDurationEnv("TEST_PARSE_DUR_ZERO", time.Minute); d != 0 {
		t.Fatalf("zero duration got %v", d)
	}
}

func TestLoad_DBMaxIdleCappedByMaxOpen(t *testing.T) {
	_ = os.Unsetenv("DB_MAX_OPEN_CONNS")
	_ = os.Unsetenv("DB_MAX_IDLE_CONNS")
	t.Setenv("DB_MAX_OPEN_CONNS", "10")
	t.Setenv("DB_MAX_IDLE_CONNS", "50")
	cfg := Load()
	if cfg.DBMaxIdleConns != 10 {
		t.Fatalf("DBMaxIdleConns=%d want 10 (harus tidak melebihi max open)", cfg.DBMaxIdleConns)
	}
	if cfg.DBMaxOpenConns != 10 {
		t.Fatalf("DBMaxOpenConns=%d want 10", cfg.DBMaxOpenConns)
	}
}
