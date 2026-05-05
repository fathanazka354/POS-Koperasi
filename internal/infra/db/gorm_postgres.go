package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/fathanazka354/pos-koperasi/internal/config"
)

// OpenGormPostgres membuka koneksi GORM ke Postgres menggunakan config yang sama.
func OpenGormPostgres(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	g, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Default logger terlalu noisy untuk API; gunakan Warn.
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := g.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	// Ping untuk memastikan koneksi valid (mirip sqlx.Connect).
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return g, nil
}

// CloseGorm menutup underlying *sql.DB dari *gorm.DB.
func CloseGorm(g *gorm.DB) error {
	if g == nil {
		return nil
	}
	sqlDB, err := g.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DBFromGorm mengembalikan underlying *sql.DB bila diperlukan.
func DBFromGorm(g *gorm.DB) (*sql.DB, error) {
	if g == nil {
		return nil, fmt.Errorf("gorm db is nil")
	}
	return g.DB()
}

