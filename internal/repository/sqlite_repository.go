package repository

import (
	"log"

	"github.com/glebarez/sqlite"
	"github.com/kouleen/lets-encrypt/internal/modle"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var sqlDb *gorm.DB

func init() {
	// 连接 SQLite（文件不存在会自动创建）
	db, err := gorm.Open(sqlite.Open("./lets-encrypt.db?_loc=Local&parseTime=true&_journal_mode=WAL&_cache_size=-20000"), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: true,
	})
	if err != nil {
		log.Fatal("sqlite 连接失败: %w", err)
	}

	// 获取底层 DB 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取 DB 失败: %w", err)
	}

	// SQLite 连接池配置（轻量即可）
	sqlDB.SetMaxOpenConns(1)    // 最大打开连接
	sqlDB.SetMaxIdleConns(1)    // 最大空闲连接
	sqlDB.SetConnMaxLifetime(0) // 连接永不过期

	// 赋值给外部指针
	sqlDb = db
	err = sqlDb.AutoMigrate(&modle.AcmeAccount{}, &modle.AcmeEncrypt{})
	if err != nil {
		log.Fatal("初始化 DB 失败: %w", err)
	}
	log.Println("✅ SQLite 连接成功")
}

func GetSqliteDb() *gorm.DB {
	return sqlDb
}
