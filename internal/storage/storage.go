package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// 数据库文件路径
	DBPath = "./data/cookiecloud.db"
)

// CookieData Cookie数据结构
type CookieData struct {
	Encrypted string `json:"encrypted"`
}

// 全局数据库连接
var db *sql.DB

// InitDataDir 初始化数据目录和数据库
func InitDataDir() error {
	// 创建数据目录
	dataDir := filepath.Dir(DBPath)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	// 连接数据库
	var err error
	db, err = sql.Open("sqlite3", DBPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 创建表
	return createTableIfNotExists()
}

// createTableIfNotExists 创建数据表（如果不存在）
func createTableIfNotExists() error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS cookies (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TRIGGER IF NOT EXISTS update_updated_at
	AFTER UPDATE ON cookies
	FOR EACH ROW
	BEGIN
		UPDATE cookies SET updated_at = CURRENT_TIMESTAMP WHERE uuid = NEW.uuid;
	END;
	`

	_, err := db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// SaveEncryptedData 保存加密数据到数据库中
func SaveEncryptedData(uuid, encrypted string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	// 使用预处理语句防止SQL注入
	stmt, err := db.Prepare(`INSERT OR REPLACE INTO cookies (uuid, encrypted) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// 执行插入操作
	_, err = stmt.Exec(uuid, encrypted)
	return err
}

// LoadEncryptedData 从数据库中加载加密数据
func LoadEncryptedData(uuid string) (*CookieData, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var encrypted string

	// 查询数据
	err := db.QueryRow("SELECT encrypted FROM cookies WHERE uuid = ?", uuid).Scan(&encrypted)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到数据，返回错误
			return nil, fmt.Errorf("no data found for uuid: %s", uuid)
		}
		return nil, fmt.Errorf("failed to load encrypted data: %w", err)
	}

	// 返回CookieData结构体
	return &CookieData{
		Encrypted: encrypted,
	}, nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}