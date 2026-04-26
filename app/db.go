package app

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	gormDB *gorm.DB
	dbPath string
	dbOnce sync.Once
	dbErr  error
)

func InitDB() error {
	dbOnce.Do(func() {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, ".cf-ssl-manager")
		}
		dbDir := filepath.Join(appData, "cf-ssl-manager")
		fmt.Printf("[InitDB] dbDir=%q\n", dbDir)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			dbErr = fmt.Errorf("创建数据目录失败: %v", err)
			fmt.Printf("[InitDB] MkdirAll failed: %v\n", err)
			return
		}

		dbPath = filepath.Join(dbDir, "certs.db")
		fmt.Printf("[InitDB] dbPath=%q\n", dbPath)

		gormDB, dbErr = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Warn),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if dbErr != nil {
			dbErr = fmt.Errorf("打开数据库失败: %v", dbErr)
			fmt.Printf("[InitDB] gorm.Open failed: %v\n", dbErr)
			return
		}

		// Enable WAL mode for better concurrent access on Windows
		if ret := gormDB.Exec("PRAGMA journal_mode=WAL"); ret.Error != nil {
			fmt.Printf("[InitDB] PRAGMA journal_mode=WAL failed: %v\n", ret.Error)
		} else {
			fmt.Printf("[InitDB] PRAGMA journal_mode=WAL OK\n")
		}
		if ret := gormDB.Exec("PRAGMA busy_timeout=5000"); ret.Error != nil {
			fmt.Printf("[InitDB] PRAGMA busy_timeout failed: %v\n", ret.Error)
		}

		if err := gormDB.AutoMigrate(&Account{}, &Zone{}, &Certificate{}, &Setting{}); err != nil {
			dbErr = fmt.Errorf("数据库迁移失败: %v", err)
			fmt.Printf("[InitDB] AutoMigrate failed: %v\n", err)
			return
		}
		fmt.Printf("[InitDB] AutoMigrate OK\n")

		// Force WAL checkpoint to ensure tables are persisted
		if ret := gormDB.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); ret.Error != nil {
			fmt.Printf("[InitDB] PRAGMA wal_checkpoint failed: %v\n", ret.Error)
		}

		fmt.Printf("[InitDB] init complete, gormDB=%v\n", gormDB != nil)
	})
	return dbErr
}

func GetDB() *gorm.DB {
	return gormDB
}

func GetDBPath() string {
	return dbPath
}

func TestDB() error {
	if gormDB == nil {
		return fmt.Errorf("数据库连接未初始化")
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %v", err)
	}
	return sqlDB.Ping()
}

func getSQLDB() (*sql.DB, error) {
	if gormDB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}
	return gormDB.DB()
}
