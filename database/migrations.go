package database

import "fmt"

// columnExists 检查表中是否存在指定列
func columnExists(db *DB, table, column string) bool {
	var count int
	// pragma_table_info 返回表的列信息
	row := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?;", table, column)
	if err := row.Scan(&count); err != nil {
		return false
	}
	return count > 0
}

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
	// 注意：SQLite ALTER TABLE ADD COLUMN 不支持 IF NOT EXISTS（部分版本），
	// 使用 pragma_table_info 先检查列是否存在，避免错误
	if !columnExists(db, "history", "entry_time") {
		_, err = tx.Exec(`ALTER TABLE history ADD COLUMN entry_time TEXT NOT NULL DEFAULT '';`)
		if err != nil {
			return fmt.Errorf("add column entry_time: %w", err)
		}
	}
	if !columnExists(db, "history", "exit_time") {
		_, err = tx.Exec(`ALTER TABLE history ADD COLUMN exit_time TEXT NOT NULL DEFAULT '';`)
		if err != nil {
			return fmt.Errorf("add column exit_time: %w", err)
		}
	}
	if !columnExists(db, "history", "holding_duration") {
		_, err = tx.Exec(`ALTER TABLE history ADD COLUMN holding_duration TEXT NOT NULL DEFAULT '';`)
		if err != nil {
			return fmt.Errorf("add column holding_duration: %w", err)
		}
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
