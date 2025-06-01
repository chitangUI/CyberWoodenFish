package database

import (
	"electronic-muyu-backend/internal/config"
	"electronic-muyu-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// 获取底层的 sql.DB 实例来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Score{},
		&models.GameSession{},
		&models.RefreshToken{},
	)
}

// InitDB 为 FX 依赖注入初始化数据库连接
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := Connect(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// 自动迁移
	if err := Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}
