package repository

import (
	"context"
	"errors"
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

func getSqliteDb() *gorm.DB {
	return sqlDb
}

func GetAcmeAccountByUsername(ctx context.Context, username string) (*modle.AcmeAccount, error) {
	acmeAccount := new(modle.AcmeAccount)
	if err := getSqliteDb().WithContext(ctx).Where("username = ?", username).First(acmeAccount).Error; err != nil {
		return nil, err
	}
	return acmeAccount, nil
}

func CreateAcmeAccount(ctx context.Context, acmeAccount *modle.AcmeAccount) (any, error) {
	if err := getSqliteDb().WithContext(ctx).Model(&modle.AcmeAccount{}).Create(acmeAccount).Error; err != nil {
		return nil, err
	}
	return acmeAccount, nil
}

func GetAcmeEncryptByDomain(ctx context.Context, domain string) (*modle.AcmeEncrypt, error) {
	acmeEncrypt := new(modle.AcmeEncrypt)
	if err := getSqliteDb().WithContext(ctx).Where("domain = ?", domain).First(acmeEncrypt).Error; err != nil {
		return nil, err
	}
	return acmeEncrypt, nil
}

func PageAcmeEncrypt(ctx context.Context, req *modle.AcmeEncryptQuery) ([]modle.AcmeEncrypt, int64, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, 0, errors.New("invalid Token")
	}
	var list []modle.AcmeEncrypt
	query := getSqliteDb().WithContext(ctx).Model(&modle.AcmeEncrypt{}).Where("username = ?", username)
	if req.Domain != "" {
		query = query.Where("domain like ?", "%"+req.Domain+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = query.Order("create_time desc")
	if err := query.Offset((req.Current - 1) * req.Size).Limit(req.Size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func CreateAcmeEncrypt(ctx context.Context, acmeEncrypt *modle.AcmeEncrypt) (any, error) {
	if err := getSqliteDb().WithContext(ctx).Create(acmeEncrypt).Error; err != nil {
		return nil, err
	}
	return acmeEncrypt, nil
}

func UpdateAcmeEncrypt(ctx context.Context, acmeEncrypt *modle.AcmeEncrypt) (any, error) {
	if err := getSqliteDb().WithContext(ctx).Where("domain = ?", acmeEncrypt.Domain).Updates(acmeEncrypt).Error; err != nil {
		return nil, err
	}
	return acmeEncrypt, nil
}

func DeleteAcmeEncryptByDomain(ctx context.Context, domain string) error {
	return getSqliteDb().WithContext(ctx).Where("domain = ?", domain).Delete(&modle.AcmeEncrypt{}).Error
}
