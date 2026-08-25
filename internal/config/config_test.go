package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfigLimits(t *testing.T) {
	cfg := Default()
	if cfg.ListenAddr != ":7700" {
		t.Fatalf("expected default ListenAddr ':7700', got %q", cfg.ListenAddr)
	}
	if cfg.MaxFileSize != 10<<30 {
		t.Fatalf("expected default MaxFileSize 10 GiB (%d bytes), got %d", int64(10<<30), cfg.MaxFileSize)
	}
	if cfg.MaxRoomSize != 10<<30 {
		t.Fatalf("expected default MaxRoomSize 10 GiB (%d bytes), got %d", int64(10<<30), cfg.MaxRoomSize)
	}
}

func TestLoadFromEnvParsesConfiguration(t *testing.T) {
	t.Setenv("LAN_DROP_DATA_DIR", t.TempDir())
	t.Setenv("LAN_DROP_MAX_FILE_SIZE", "64MiB")
	t.Setenv("LAN_DROP_DEFAULT_TTL", "30m")
	t.Setenv("LAN_DROP_CLEANUP_BATCH_SIZE", "25")
	t.Setenv("LAN_DROP_STAGING_MAX_AGE", "20m")
	t.Setenv("LAN_DROP_ORPHAN_GRACE_PERIOD", "5m")
	t.Setenv("LAN_DROP_CLOSED_ROOM_RETENTION", "10s")
	t.Setenv("LAN_DROP_GLOBAL_SHARE_ENABLED", "true")
	t.Setenv("LAN_DROP_PUBLIC_BASE_URL", "https://share.example.com")
	t.Setenv("LAN_DROP_MAX_SHARE_TTL", "12h")
	t.Setenv("LAN_DROP_DEFAULT_SHARE_TTL", "2h")
	t.Setenv("LAN_DROP_MAX_SHARES_PER_ROOM", "5")
	t.Setenv("LAN_DROP_SHARE_MANAGEMENT_RATE_LIMIT", "20")
	t.Setenv("LAN_DROP_SHARE_ACCESS_RATE_LIMIT", "200")
	t.Setenv("LAN_DROP_TRUSTED_PROXIES", "10.0.0.0/8, 192.168.0.0/16")
	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxFileSize != 64<<20 {
		t.Fatalf("expected 64 MiB, got %d", cfg.MaxFileSize)
	}
	if cfg.DefaultTTL != 30*time.Minute {
		t.Fatalf("expected 30m, got %s", cfg.DefaultTTL)
	}
	if cfg.CleanupBatchSize != 25 {
		t.Fatalf("expected batch size 25, got %d", cfg.CleanupBatchSize)
	}
	if cfg.StagingMaxAge != 20*time.Minute {
		t.Fatalf("expected staging max age 20m, got %s", cfg.StagingMaxAge)
	}
	if cfg.OrphanGracePeriod != 5*time.Minute {
		t.Fatalf("expected orphan grace period 5m, got %s", cfg.OrphanGracePeriod)
	}
	if cfg.ClosedRoomRetention != 10*time.Second {
		t.Fatalf("expected closed room retention 10s, got %s", cfg.ClosedRoomRetention)
	}
	if !cfg.GlobalShareEnabled {
		t.Fatal("expected GlobalShareEnabled to be true")
	}
	if cfg.PublicBaseURL != "https://share.example.com" {
		t.Fatalf("expected https://share.example.com, got %q", cfg.PublicBaseURL)
	}
	if cfg.MaxShareTTL != 12*time.Hour {
		t.Fatalf("expected 12h max share ttl, got %s", cfg.MaxShareTTL)
	}
	if cfg.DefaultShareTTL != 2*time.Hour {
		t.Fatalf("expected 2h default share ttl, got %s", cfg.DefaultShareTTL)
	}
	if cfg.MaxSharesPerRoom != 5 {
		t.Fatalf("expected max shares 5, got %d", cfg.MaxSharesPerRoom)
	}
	if cfg.ShareManagementRateLimit != 20 {
		t.Fatalf("expected share management rate limit 20, got %d", cfg.ShareManagementRateLimit)
	}
	if cfg.ShareAccessRateLimit != 200 {
		t.Fatalf("expected share access rate limit 200, got %d", cfg.ShareAccessRateLimit)
	}
	if len(cfg.TrustedProxies) != 2 {
		t.Fatalf("expected two proxies, got %v", cfg.TrustedProxies)
	}
}

func TestValidateRejectsInvalidTTLOrder(t *testing.T) {
	cfg := Default()
	cfg.MinTTL = time.Hour
	cfg.DefaultTTL = 30 * time.Minute
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateRejectsInvalidLogFormat(t *testing.T) {
	cfg := Default()
	cfg.LogFormat = "yaml"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidatePublicBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid https url", "https://share.example.com", false},
		{"valid https url with port", "https://share.example.com:8443", false},
		{"valid https url with path", "https://share.example.com/custom/prefix", false},
		{"reject http url", "http://share.example.com", true},
		{"reject empty url", "", true},
		{"reject url with credentials", "https://user:pass@share.example.com", true},
		{"reject url with relative path traversal", "https://share.example.com/../escaped", true},
		{"reject invalid scheme", "ftp://share.example.com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePublicBaseURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePublicBaseURL(%q) err = %v, wantErr = %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestValidateRejectsInvalidShareTTLOrder(t *testing.T) {
	cfg := Default()
	cfg.DefaultShareTTL = 2 * time.Hour
	cfg.MaxShareTTL = time.Hour
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for DefaultShareTTL > MaxShareTTL")
	}
}

func TestValidateDBPathEnforcesDataDirContainment(t *testing.T) {
	dataDir := t.TempDir()
	tests := []struct {
		name    string
		dataDir string
		dbPath  string
		wantErr bool
	}{
		{
			name:    "valid db inside data dir",
			dataDir: dataDir,
			dbPath:  filepath.Join(dataDir, "lan-drop.db"),
			wantErr: false,
		},
		{
			name:    "valid db in subfolder of data dir",
			dataDir: dataDir,
			dbPath:  filepath.Join(dataDir, "nested", "lan-drop.db"),
			wantErr: false,
		},
		{
			name:    "db path escaping data dir via dot dot",
			dataDir: dataDir,
			dbPath:  filepath.Join(dataDir, "..", "escaped.db"),
			wantErr: true,
		},
		{
			name:    "db path equals data dir",
			dataDir: dataDir,
			dbPath:  dataDir,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.DataDir = tt.dataDir
			cfg.DBPath = tt.dbPath
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromEnvHamalPrefix(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("HAMAL_LISTEN_ADDR", ":8800")
	t.Setenv("HAMAL_DATA_DIR", dataDir)
	t.Setenv("HAMAL_MAX_FILE_SIZE", "128MiB")
	t.Setenv("HAMAL_DEFAULT_TTL", "45m")
	t.Setenv("HAMAL_CLEANUP_BATCH_SIZE", "30")
	t.Setenv("HAMAL_LOG_LEVEL", "debug")
	t.Setenv("HAMAL_LOG_FORMAT", "text")
	t.Setenv("HAMAL_TRUSTED_PROXIES", "172.16.0.0/12")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv failed: %v", err)
	}
	if cfg.ListenAddr != ":8800" {
		t.Fatalf("expected ListenAddr ':8800', got %q", cfg.ListenAddr)
	}
	if cfg.DataDir != dataDir {
		t.Fatalf("expected DataDir %q, got %q", dataDir, cfg.DataDir)
	}
	if cfg.MaxFileSize != 128<<20 {
		t.Fatalf("expected MaxFileSize 128 MiB, got %d", cfg.MaxFileSize)
	}
	if cfg.DefaultTTL != 45*time.Minute {
		t.Fatalf("expected DefaultTTL 45m, got %s", cfg.DefaultTTL)
	}
	if cfg.CleanupBatchSize != 30 {
		t.Fatalf("expected CleanupBatchSize 30, got %d", cfg.CleanupBatchSize)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("expected LogLevel 'debug', got %q", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Fatalf("expected LogFormat 'text', got %q", cfg.LogFormat)
	}
	if len(cfg.TrustedProxies) != 1 || cfg.TrustedProxies[0] != "172.16.0.0/12" {
		t.Fatalf("expected TrustedProxies ['172.16.0.0/12'], got %v", cfg.TrustedProxies)
	}
}

func TestLoadFromEnvPrecedence(t *testing.T) {
	dataDir := t.TempDir()
	// When both HAMAL_* and LAN_DROP_* are present, HAMAL_* must take precedence
	t.Setenv("HAMAL_DATA_DIR", dataDir)
	t.Setenv("LAN_DROP_DATA_DIR", "/nonexistent/data/dir")

	t.Setenv("HAMAL_LISTEN_ADDR", ":7701")
	t.Setenv("LAN_DROP_LISTEN_ADDR", ":9999")

	t.Setenv("HAMAL_MAX_FILE_SIZE", "2GiB")
	t.Setenv("LAN_DROP_MAX_FILE_SIZE", "500MiB")

	t.Setenv("HAMAL_LOG_LEVEL", "warn")
	t.Setenv("LAN_DROP_LOG_LEVEL", "error")

	// Test fallback when only LAN_DROP_* is provided
	t.Setenv("LAN_DROP_DEFAULT_TTL", "2h")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv failed: %v", err)
	}
	if cfg.ListenAddr != ":7701" {
		t.Fatalf("expected HAMAL_LISTEN_ADDR ':7701' to take precedence over ':9999', got %q", cfg.ListenAddr)
	}
	if cfg.DataDir != dataDir {
		t.Fatalf("expected HAMAL_DATA_DIR %q to take precedence, got %q", dataDir, cfg.DataDir)
	}
	if cfg.MaxFileSize != 2<<30 {
		t.Fatalf("expected HAMAL_MAX_FILE_SIZE 2 GiB to take precedence, got %d", cfg.MaxFileSize)
	}
	if cfg.LogLevel != "warn" {
		t.Fatalf("expected HAMAL_LOG_LEVEL 'warn' to take precedence, got %q", cfg.LogLevel)
	}
	if cfg.DefaultTTL != 2*time.Hour {
		t.Fatalf("expected fallback LAN_DROP_DEFAULT_TTL '2h', got %s", cfg.DefaultTTL)
	}
}
