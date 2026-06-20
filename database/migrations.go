package database

import "fmt"

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

	// 2. 创建 history 表（已调出）
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id               INTEGER PRIMARY KEY AUTOINCREMENT,
			code             TEXT    NOT NULL,
			name             TEXT    NOT NULL,
			entry_date       TEXT    NOT NULL,
			entry_price      INTEGER NOT NULL,
			exit_date        TEXT    NOT NULL,
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

	// 2.5 迁移：为旧表添加 holding_duration 列（如果不存在）
	_, _ = tx.Exec(`ALTER TABLE history ADD COLUMN IF NOT EXISTS holding_duration TEXT NOT NULL DEFAULT '';`)

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
