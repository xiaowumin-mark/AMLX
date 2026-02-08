package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	driver "github.com/go-sql-driver/mysql"
	"github.com/xiaowumin-mark/AMLX/config"
	"github.com/xiaowumin-mark/AMLX/logx"
	"github.com/xiaowumin-mark/AMLX/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg config.MySQLConfig) (*gorm.DB, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}
	gormConfig := &gorm.Config{
		Logger: newGormLogger(cfg),
	}
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Users{},
		&model.Roles{},
		&model.Permissions{},
		&model.RolePermissions{},
		&model.RefreshTokens{},
	)
}

func buildDSN(cfg config.MySQLConfig) (string, error) {
	loc, err := time.LoadLocation(cfg.Loc)
	if err != nil {
		return "", fmt.Errorf("mysql.loc: %w", err)
	}
	dsnCfg := driver.Config{
		User:                 cfg.User,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:               cfg.Database,
		ParseTime:            cfg.ParseTimeValue(),
		Loc:                  loc,
		AllowNativePasswords: true,
	}
	dsnCfg.Params = map[string]string{
		"charset": cfg.Charset,
	}
	return dsnCfg.FormatDSN(), nil
}

func newGormLogger(cfg config.MySQLConfig) logger.Interface {
	level := logger.Warn
	switch strings.ToLower(cfg.LogLevel) {
	case "silent":
		level = logger.Silent
	case "error":
		level = logger.Error
	case "info":
		level = logger.Info
	case "warn", "warning":
		level = logger.Warn
	}

	return logger.New(
		log.New(logx.Writer(), "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(cfg.SlowThresholdMs) * time.Millisecond,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}
