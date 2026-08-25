package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr               string
	BaseURL                  string
	DataDir                  string
	DBPath                   string
	ServerSecret             string
	MaxTTL                   time.Duration
	DefaultTTL               time.Duration
	MinTTL                   time.Duration
	MaxFileSize              int64
	MaxRoomSize              int64
	MaxFilesPerRoom          int
	MaxTotalStorage          int64
	MinFreeSpace             int64
	CleanupInterval          time.Duration
	CleanupBatchSize         int
	StagingMaxAge            time.Duration
	OrphanGracePeriod        time.Duration
	ClosedRoomRetention      time.Duration
	GlobalShareEnabled       bool
	PublicBaseURL            string
	MaxShareTTL              time.Duration
	DefaultShareTTL          time.Duration
	MaxSharesPerRoom         int
	ShareManagementRateLimit int
	ShareAccessRateLimit     int
	UploadIdleTimeout        time.Duration
	LogFormat                string
	LogLevel                 string
	TrustedProxies           []string
	SecureCookies            string
}

func Default() Config {
	dataDir := "/data"
	return Config{
		ListenAddr: ":7700", DataDir: dataDir, DBPath: filepath.Join(dataDir, "lan-drop.db"),
		MaxTTL: 24 * time.Hour, DefaultTTL: time.Hour, MinTTL: 5 * time.Minute,
		MaxFileSize: 10 << 30, MaxRoomSize: 10 << 30, MaxFilesPerRoom: 100,
		MaxTotalStorage: 100 << 30, MinFreeSpace: 5 << 30,
		CleanupInterval: time.Minute, CleanupBatchSize: 50,
		StagingMaxAge: 15 * time.Minute, OrphanGracePeriod: 10 * time.Minute, ClosedRoomRetention: 0,
		GlobalShareEnabled: false, PublicBaseURL: "",
		MaxShareTTL: 24 * time.Hour, DefaultShareTTL: time.Hour,
		MaxSharesPerRoom: 10, ShareManagementRateLimit: 30, ShareAccessRateLimit: 300,
		UploadIdleTimeout: 10 * time.Minute,
		LogFormat:         "json", LogLevel: "info", SecureCookies: "auto",
	}
}

func LoadFromEnv() (Config, error) {
	c := Default()
	c.ListenAddr = envString(c.ListenAddr, "HAMAL_LISTEN_ADDR", "LAN_DROP_LISTEN_ADDR")
	c.BaseURL = envString(c.BaseURL, "HAMAL_BASE_URL", "LAN_DROP_BASE_URL")
	c.DataDir = envString(c.DataDir, "HAMAL_DATA_DIR", "LAN_DROP_DATA_DIR")
	c.DBPath = envString(filepath.Join(c.DataDir, "lan-drop.db"), "HAMAL_DB_PATH", "LAN_DROP_DB_PATH")
	c.ServerSecret = strings.TrimSpace(envString("", "HAMAL_SERVER_SECRET", "LAN_DROP_SERVER_SECRET"))
	c.LogFormat = strings.ToLower(envString(c.LogFormat, "HAMAL_LOG_FORMAT", "LAN_DROP_LOG_FORMAT"))
	c.LogLevel = strings.ToLower(envString(c.LogLevel, "HAMAL_LOG_LEVEL", "LAN_DROP_LOG_LEVEL"))
	c.SecureCookies = strings.ToLower(envString(c.SecureCookies, "HAMAL_SECURE_COOKIES", "LAN_DROP_SECURE_COOKIES"))
	c.TrustedProxies = splitCSV(envString("", "HAMAL_TRUSTED_PROXIES", "LAN_DROP_TRUSTED_PROXIES"))

	var err error
	if c.MaxTTL, err = envDuration(c.MaxTTL, "HAMAL_MAX_TTL", "LAN_DROP_MAX_TTL"); err != nil {
		return c, err
	}
	if c.DefaultTTL, err = envDuration(c.DefaultTTL, "HAMAL_DEFAULT_TTL", "LAN_DROP_DEFAULT_TTL"); err != nil {
		return c, err
	}
	if c.MinTTL, err = envDuration(c.MinTTL, "HAMAL_MIN_TTL", "LAN_DROP_MIN_TTL"); err != nil {
		return c, err
	}
	if c.CleanupInterval, err = envDuration(c.CleanupInterval, "HAMAL_CLEANUP_INTERVAL", "LAN_DROP_CLEANUP_INTERVAL"); err != nil {
		return c, err
	}
	if c.CleanupBatchSize, err = envInt(c.CleanupBatchSize, "HAMAL_CLEANUP_BATCH_SIZE", "LAN_DROP_CLEANUP_BATCH_SIZE"); err != nil {
		return c, err
	}
	if c.StagingMaxAge, err = envDuration(c.StagingMaxAge, "HAMAL_STAGING_MAX_AGE", "LAN_DROP_STAGING_MAX_AGE"); err != nil {
		return c, err
	}
	if c.OrphanGracePeriod, err = envDuration(c.OrphanGracePeriod, "HAMAL_ORPHAN_GRACE_PERIOD", "LAN_DROP_ORPHAN_GRACE_PERIOD"); err != nil {
		return c, err
	}
	if c.ClosedRoomRetention, err = envDuration(c.ClosedRoomRetention, "HAMAL_CLOSED_ROOM_RETENTION", "LAN_DROP_CLOSED_ROOM_RETENTION"); err != nil {
		return c, err
	}
	if c.GlobalShareEnabled, err = envBool(c.GlobalShareEnabled, "HAMAL_GLOBAL_SHARE_ENABLED", "LAN_DROP_GLOBAL_SHARE_ENABLED"); err != nil {
		return c, err
	}
	c.PublicBaseURL = strings.TrimSpace(envString(c.PublicBaseURL, "HAMAL_PUBLIC_BASE_URL", "LAN_DROP_PUBLIC_BASE_URL"))
	if c.MaxShareTTL, err = envDuration(c.MaxShareTTL, "HAMAL_MAX_SHARE_TTL", "LAN_DROP_MAX_SHARE_TTL"); err != nil {
		return c, err
	}
	if c.DefaultShareTTL, err = envDuration(c.DefaultShareTTL, "HAMAL_DEFAULT_SHARE_TTL", "LAN_DROP_DEFAULT_SHARE_TTL"); err != nil {
		return c, err
	}
	if c.MaxSharesPerRoom, err = envInt(c.MaxSharesPerRoom, "HAMAL_MAX_SHARES_PER_ROOM", "LAN_DROP_MAX_SHARES_PER_ROOM"); err != nil {
		return c, err
	}
	if c.ShareManagementRateLimit, err = envInt(c.ShareManagementRateLimit, "HAMAL_SHARE_MANAGEMENT_RATE_LIMIT", "LAN_DROP_SHARE_MANAGEMENT_RATE_LIMIT"); err != nil {
		return c, err
	}
	if c.ShareAccessRateLimit, err = envInt(c.ShareAccessRateLimit, "HAMAL_SHARE_ACCESS_RATE_LIMIT", "LAN_DROP_SHARE_ACCESS_RATE_LIMIT"); err != nil {
		return c, err
	}
	if c.UploadIdleTimeout, err = envDuration(c.UploadIdleTimeout, "HAMAL_UPLOAD_IDLE_TIMEOUT", "LAN_DROP_UPLOAD_IDLE_TIMEOUT"); err != nil {
		return c, err
	}
	if c.MaxFileSize, err = envBytes(c.MaxFileSize, "HAMAL_MAX_FILE_SIZE", "LAN_DROP_MAX_FILE_SIZE"); err != nil {
		return c, err
	}
	if c.MaxRoomSize, err = envBytes(c.MaxRoomSize, "HAMAL_MAX_ROOM_SIZE", "LAN_DROP_MAX_ROOM_SIZE"); err != nil {
		return c, err
	}
	if c.MaxTotalStorage, err = envBytes(c.MaxTotalStorage, "HAMAL_MAX_TOTAL_STORAGE", "LAN_DROP_MAX_TOTAL_STORAGE"); err != nil {
		return c, err
	}
	if c.MinFreeSpace, err = envBytes(c.MinFreeSpace, "HAMAL_MIN_FREE_SPACE", "LAN_DROP_MIN_FREE_SPACE"); err != nil {
		return c, err
	}
	if c.MaxFilesPerRoom, err = envInt(c.MaxFilesPerRoom, "HAMAL_MAX_FILES_PER_ROOM", "LAN_DROP_MAX_FILES_PER_ROOM"); err != nil {
		return c, err
	}
	return c, c.Validate()
}

func ValidatePublicBaseURL(urlStr string) error {
	trimmed := strings.TrimSpace(urlStr)
	if trimmed == "" {
		return fmt.Errorf("public base URL cannot be empty")
	}
	u, err := url.ParseRequestURI(trimmed)
	if err != nil {
		return fmt.Errorf("invalid public base URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("LAN_DROP_PUBLIC_BASE_URL must use https scheme (got %q)", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("LAN_DROP_PUBLIC_BASE_URL must specify a hostname")
	}
	if u.User != nil {
		return fmt.Errorf("LAN_DROP_PUBLIC_BASE_URL must not contain user credentials")
	}
	if strings.Contains(u.Path, "..") {
		return fmt.Errorf("LAN_DROP_PUBLIC_BASE_URL path must not contain relative traversal")
	}
	return nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.ListenAddr) == "" {
		return fmt.Errorf("LAN_DROP_LISTEN_ADDR cannot be empty")
	}
	if strings.TrimSpace(c.DataDir) == "" || strings.TrimSpace(c.DBPath) == "" {
		return fmt.Errorf("data directory and database path are required")
	}
	absData, err := filepath.Abs(c.DataDir)
	if err != nil {
		return fmt.Errorf("invalid data directory: %w", err)
	}
	absDB, err := filepath.Abs(c.DBPath)
	if err != nil {
		return fmt.Errorf("invalid database path: %w", err)
	}
	rel, err := filepath.Rel(absData, absDB)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("database path %q must be located within data directory %q", c.DBPath, c.DataDir)
	}
	if c.BaseURL != "" {
		u, err := url.ParseRequestURI(c.BaseURL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("LAN_DROP_BASE_URL must be an absolute URL")
		}
	}
	if c.PublicBaseURL != "" {
		if err := ValidatePublicBaseURL(c.PublicBaseURL); err != nil {
			return err
		}
	}
	if len(c.ServerSecret) > 0 && len(c.ServerSecret) < 32 {
		return fmt.Errorf("LAN_DROP_SERVER_SECRET must be at least 32 bytes")
	}
	if c.MinTTL <= 0 || c.DefaultTTL < c.MinTTL || c.MaxTTL < c.DefaultTTL {
		return fmt.Errorf("TTL values must satisfy 0 < minimum <= default <= maximum")
	}
	if c.MaxShareTTL <= 0 || c.DefaultShareTTL <= 0 || c.MaxShareTTL < c.DefaultShareTTL {
		return fmt.Errorf("share TTL values must satisfy 0 < default <= maximum")
	}
	if c.MaxFileSize <= 0 || c.MaxRoomSize <= 0 || c.MaxTotalStorage <= 0 || c.MinFreeSpace < 0 || c.MaxFilesPerRoom <= 0 {
		return fmt.Errorf("storage limits must be positive (minimum free space may be zero)")
	}
	if c.CleanupInterval <= 0 || c.CleanupBatchSize <= 0 || c.StagingMaxAge <= 0 || c.OrphanGracePeriod <= 0 || c.ClosedRoomRetention < 0 || c.UploadIdleTimeout <= 0 {
		return fmt.Errorf("cleanup parameters and upload idle timeout must be positive (closed room retention may be zero)")
	}
	if c.MaxSharesPerRoom <= 0 || c.ShareManagementRateLimit <= 0 || c.ShareAccessRateLimit <= 0 {
		return fmt.Errorf("share limits and rate limits must be positive")
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("LAN_DROP_LOG_FORMAT must be json or text")
	}
	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return fmt.Errorf("LAN_DROP_LOG_LEVEL must be debug, info, warn, or error")
	}
	if c.SecureCookies != "auto" && c.SecureCookies != "true" && c.SecureCookies != "false" {
		return fmt.Errorf("LAN_DROP_SECURE_COOKIES must be auto, true, or false")
	}
	for _, proxy := range c.TrustedProxies {
		if _, _, err := net.ParseCIDR(proxy); err != nil {
			return fmt.Errorf("invalid trusted proxy CIDR %q", proxy)
		}
	}
	return nil
}

func lookupEnvFirst(keys ...string) (string, string) {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			return val, key
		}
	}
	if len(keys) > 0 {
		return "", keys[0]
	}
	return "", ""
}

func envString(fallback string, keys ...string) string {
	if val, _ := lookupEnvFirst(keys...); val != "" {
		return val
	}
	return fallback
}

func envDuration(fallback time.Duration, keys ...string) (time.Duration, error) {
	val, key := lookupEnvFirst(keys...)
	if val == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}

func envInt(fallback int, keys ...string) (int, error) {
	val, key := lookupEnvFirst(keys...)
	if val == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return parsed, nil
}

func envBytes(fallback int64, keys ...string) (int64, error) {
	val, key := lookupEnvFirst(keys...)
	if val == "" {
		return fallback, nil
	}
	multiplier := int64(1)
	lower := strings.ToLower(val)
	units := []struct {
		suffix string
		factor int64
	}{
		{"gib", 1 << 30}, {"mib", 1 << 20}, {"kib", 1 << 10},
		{"gb", 1 << 30}, {"mb", 1 << 20}, {"kb", 1 << 10}, {"b", 1},
	}
	for _, unit := range units {
		if strings.HasSuffix(lower, unit.suffix) {
			multiplier = unit.factor
			lower = strings.TrimSpace(strings.TrimSuffix(lower, unit.suffix))
			break
		}
	}
	parsed, err := strconv.ParseInt(lower, 10, 64)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%s must be a non-negative byte value", key)
	}
	return parsed * multiplier, nil
}

func splitCSV(value string) []string {
	var result []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func envBool(fallback bool, keys ...string) (bool, error) {
	val, key := lookupEnvFirst(keys...)
	if val == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false: %w", key, err)
	}
	return parsed, nil
}
