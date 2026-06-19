package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB 封装 SQLite 数据库连接，提供便捷方法
type DB struct {
	conn *sql.DB
}

// Init 打开数据库连接，启用 WAL 模式，设置 busy_timeout，并执行迁移
func Init(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 基本连接验证
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db := &DB{conn: conn}

	// 启用 WAL 模式 + 设置 busy_timeout
	if err := db.setPragmas(); err != nil {
		db.Close()
		return nil, fmt.Errorf("set pragmas: %w", err)
	}

	// 执行表结构和索引迁移
	if err := db.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// setPragmas 启用 WAL 模式并设置 busy_timeout
func (db *DB) setPragmas() error {
	if _, err := db.conn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("enable WAL: %w", err)
	}
	if _, err := db.conn.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}
	return nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn 返回底层 *sql.DB（用于需要直接使用 sql.DB 的场景）
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Exec 执行 SQL（INSERT/UPDATE/DELETE），返回 sql.Result
func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// Query 执行查询（SELECT），返回 sql.Rows
func (db *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow 执行单行查询，返回 sql.Row
func (db *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// Begin 开启事务
func (db *DB) Begin() (*sql.Tx, error) {
	return db.conn.Begin()
}
