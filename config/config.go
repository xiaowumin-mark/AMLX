package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MySQL  MySQLConfig  `yaml:"mysql"`
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	Auth   AuthConfig   `yaml:"auth"`
}

type MySQLConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	Charset         string        `yaml:"charset"`
	ParseTime       *bool         `yaml:"parse_time"`
	Loc             string        `yaml:"loc"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	LogLevel        string        `yaml:"log_level"`
	SlowThresholdMs int           `yaml:"slow_threshold_ms"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	Log          bool          `yaml:"log"`
	LogLevel     string        `yaml:"log_level"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	File       string `yaml:"file"`
	AddSource  bool   `yaml:"add_source"`
	TimeFormat string `yaml:"time_format"`
}

type AuthConfig struct {
	JWTSecret          string        `yaml:"jwt_secret"`
	Issuer             string        `yaml:"issuer"`
	AccessTTL          time.Duration `yaml:"access_ttl"`
	RefreshTTL         time.Duration `yaml:"refresh_ttl"`
	BcryptCost         int           `yaml:"bcrypt_cost"`
	AllowRegister      *bool         `yaml:"allow_register"`
	AllowRegisterRole  bool          `yaml:"allow_register_role"`
	DefaultRoleID      uint          `yaml:"default_role_id"`
	RefreshTokenReuse  bool          `yaml:"refresh_token_reuse"`
	BootstrapAdminRole *bool         `yaml:"bootstrap_admin_role"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c MySQLConfig) ParseTimeValue() bool {
	if c.ParseTime == nil {
		return true
	}
	return *c.ParseTime
}

func (c AuthConfig) AllowRegisterValue() bool {
	if c.AllowRegister == nil {
		return true
	}
	return *c.AllowRegister
}

func (c AuthConfig) BootstrapAdminRoleValue() bool {
	if c.BootstrapAdminRole == nil {
		return true
	}
	return *c.BootstrapAdminRole
}

func applyDefaults(cfg *Config) {
	if cfg.MySQL.Host == "" {
		cfg.MySQL.Host = "127.0.0.1"
	}
	if cfg.MySQL.Port == 0 {
		cfg.MySQL.Port = 3306
	}
	if cfg.MySQL.Charset == "" {
		cfg.MySQL.Charset = "utf8mb4"
	}
	if cfg.MySQL.Loc == "" {
		cfg.MySQL.Loc = "Local"
	}
	if cfg.MySQL.ParseTime == nil {
		value := true
		cfg.MySQL.ParseTime = &value
	}
	if cfg.MySQL.MaxOpenConns == 0 {
		cfg.MySQL.MaxOpenConns = 25
	}
	if cfg.MySQL.MaxIdleConns == 0 {
		cfg.MySQL.MaxIdleConns = 10
	}
	if cfg.MySQL.ConnMaxLifetime == 0 {
		cfg.MySQL.ConnMaxLifetime = 30 * time.Minute
	}
	if cfg.MySQL.ConnMaxIdleTime == 0 {
		cfg.MySQL.ConnMaxIdleTime = 10 * time.Minute
	}
	if cfg.MySQL.LogLevel == "" {
		cfg.MySQL.LogLevel = "warn"
	}
	if cfg.MySQL.SlowThresholdMs == 0 {
		cfg.MySQL.SlowThresholdMs = 200
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.LogLevel == "" {
		cfg.Server.LogLevel = "info"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 5 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 10 * time.Second
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 60 * time.Second
	}

	if cfg.Log.Level == "" {
		cfg.Log.Level = cfg.Server.LogLevel
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}
	if cfg.Log.Output == "" {
		if cfg.Server.Log {
			cfg.Log.Output = "stdout"
		} else {
			cfg.Log.Output = "discard"
		}
	}
	if cfg.Log.File == "" {
		cfg.Log.File = "logs/amlx.log"
	}
	if cfg.Log.TimeFormat == "" {
		cfg.Log.TimeFormat = time.RFC3339
	}

	if cfg.Auth.Issuer == "" {
		cfg.Auth.Issuer = "AMLX"
	}
	if cfg.Auth.AccessTTL == 0 {
		cfg.Auth.AccessTTL = 15 * time.Minute
	}
	if cfg.Auth.RefreshTTL == 0 {
		cfg.Auth.RefreshTTL = 7 * 24 * time.Hour
	}
	if cfg.Auth.BcryptCost == 0 {
		cfg.Auth.BcryptCost = 10
	}
	if cfg.Auth.DefaultRoleID == 0 {
		cfg.Auth.DefaultRoleID = 1
	}
	if cfg.Auth.AllowRegister == nil {
		value := true
		cfg.Auth.AllowRegister = &value
	}
	if cfg.Auth.BootstrapAdminRole == nil {
		value := true
		cfg.Auth.BootstrapAdminRole = &value
	}
}

func validate(cfg *Config) error {
	if cfg.MySQL.Host == "" {
		return errors.New("mysql.host is required")
	}
	if cfg.MySQL.Port <= 0 {
		return errors.New("mysql.port must be greater than 0")
	}
	if cfg.MySQL.User == "" {
		return errors.New("mysql.user is required")
	}
	if cfg.MySQL.Database == "" {
		return errors.New("mysql.database is required")
	}
	if cfg.Server.Port <= 0 {
		return errors.New("server.port must be greater than 0")
	}
	switch strings.ToLower(cfg.Log.Output) {
	case "stdout", "stderr", "file", "both", "discard", "none", "off", "":
	default:
		return errors.New("log.output must be stdout|stderr|file|both|discard")
	}
	if strings.TrimSpace(cfg.Auth.JWTSecret) == "" {
		return errors.New("auth.jwt_secret is required")
	}
	return nil
}
