package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// migrate 执行数据库迁移：创建表、索引、设置 WAL 和 busy_timeout
func (db *DB) migrate() (err error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// 1. 创建 stocks 表（当前持仓）
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS stocks (
			id             INTEGER PRIMARY KEY AUTOINCREMENT,
			code           TEXT    NOT NULL UNIQUE,
			name           TEXT    NOT NULL,
			entry_date     TEXT    NOT NULL,
			entry_time     TEXT    NOT NULL,
			entry_price    INTEGER NOT NULL,
			raw_price      INTEGER NOT NULL,
			adjust_factor  INTEGER DEFAULT 1000,
			current_price  INTEGER NOT NULL,
			prev_close     INTEGER NOT NULL,
			daily_change   INTEGER DEFAULT 0,
			acc_change     INTEGER DEFAULT 0,
			data_source    TEXT    DEFAULT 'tencent',
			status         TEXT    DEFAULT 'normal',
			last_update    TEXT,
			created_at     TEXT    NOT NULL,
			updated_at     TEXT    NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("create stocks table: %w", err)
	}

	// 2. 创建 history 表（已调出）—— 使用最新完整 schema
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			code             TEXT    NOT NULL,
			name             TEXT    NOT NULL,
			entry_date       TEXT    NOT NULL,
			entry_time       TEXT    NOT NULL DEFAULT '',
			entry_price      INTEGER NOT NULL,
			exit_date        TEXT    NOT NULL,
			exit_time        TEXT    NOT NULL DEFAULT '',
			exit_price       INTEGER NOT NULL,
			holding_days     INTEGER NOT NULL,
			holding_duration TEXT    NOT NULL DEFAULT '',
			total_return     INTEGER NOT NULL,
			data_source      TEXT    DEFAULT 'tencent',
			created_at       TEXT    NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("create history table: %w", err)
	}

	// 2.5 迁移：为旧表添加新列（如果不存在）
	if err := addColumnIfMissing(tx, "history", "entry_time", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return fmt.Errorf("migrate history.entry_time: %w", err)
	}
	if err := addColumnIfMissing(tx, "history", "exit_time", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return fmt.Errorf("migrate history.exit_time: %w", err)
	}
	if err := addColumnIfMissing(tx, "history", "holding_duration", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return fmt.Errorf("migrate history.holding_duration: %w", err)
	}

	// 3. 创建 daily_snapshots 表（每日快照，级联删除）
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS daily_snapshots (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			stock_id    INTEGER NOT NULL,
			date        TEXT    NOT NULL,
			price       INTEGER NOT NULL,
			change_bp   INTEGER DEFAULT 0,
			FOREIGN KEY (stock_id) REFERENCES stocks(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		return fmt.Errorf("create daily_snapshots table: %w", err)
	}

	// 4. 创建索引
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_stocks_code ON stocks(code);`,
		`CREATE INDEX IF NOT EXISTS idx_history_code ON history(code);`,
		`CREATE INDEX IF NOT EXISTS idx_history_exit_date ON history(exit_date DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_snapshots_stock_date ON daily_snapshots(stock_id, date);`,
	}
	for _, idx := range indexes {
		_, err = tx.Exec(idx)
		if err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}

	return nil
}

// addColumnIfMissing 使用 PRAGMA table_info 检查列是否存在，不存在则添加
// 如果 ALTER TABLE 因列已存在而失败，自动忽略该错误
func addColumnIfMissing(tx *sql.Tx, table, column, def string) error {
	// 1. 通过 PRAGMA table_info 检查列是否存在
	rows, err := tx.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return fmt.Errorf("pragma table_info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, typ, notNull, dfltValue, pk string
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dfltValue, &pk); err != nil {
			continue
		}
		if name == column {
			return nil // 列已存在
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate pragma table_info: %w", err)
	}

	// 2. 列不存在，执行 ALTER TABLE ADD COLUMN
	_, err = tx.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + def + ";")
	if err != nil {
		// 如果是重复列错误，忽略
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "duplicate column") ||
			strings.Contains(errStr, "already has column") ||
			strings.Contains(errStr, "column already exists") {
			return nil
		}
		return fmt.Errorf("alter table add %s: %w", column, err)
	}
	return nil
}
