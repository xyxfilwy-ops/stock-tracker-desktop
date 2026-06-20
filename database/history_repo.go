package database

import (
	"database/sql"
	"fmt"
)

// HistoryRepository 提供 history 表的 CRUD 操作
type HistoryRepository struct {
	db *DB
}

// NewHistoryRepository 创建 HistoryRepository 实例
func NewHistoryRepository(db *DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// Create 向 history 表插入一条新记录（调出时调用）
func (r *HistoryRepository) Create(h *HistoryRecord) (*HistoryRecord, error) {
	query := `
		INSERT INTO history (
			code, name, entry_date, entry_price, exit_date, exit_price,
			holding_days, holding_duration, total_return, data_source, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id;
	`
	var id int64
	err := r.db.QueryRow(query,
		h.Code, h.Name, h.EntryDate, h.EntryPrice, h.ExitDate, h.ExitPrice,
		h.HoldingDays, h.HoldingDuration, h.TotalReturn, h.DataSource, h.CreatedAt,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert history: %w", err)
	}
	h.ID = id
	return h, nil
}

// GetAll 返回所有历史记录，按调出日期降序排列（最新的在前）
func (r *HistoryRepository) GetAll() ([]HistoryRecord, error) {
	query := `
		SELECT
			id, code, name, entry_date, entry_price, exit_date, exit_price,
			holding_days, holding_duration, total_return, data_source, created_at
		FROM history
		ORDER BY exit_date DESC, created_at DESC;
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetByCode 根据股票代码查询所有相关历史记录
func (r *HistoryRepository) GetByCode(code string) ([]HistoryRecord, error) {
	query := `
		SELECT
			id, code, name, entry_date, entry_price, exit_date, exit_price,
			holding_days, holding_duration, total_return, data_source, created_at
		FROM history
		WHERE code = ?
		ORDER BY exit_date DESC;
	`
	rows, err := r.db.Query(query, code)
	if err != nil {
		return nil, fmt.Errorf("query history by code: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetByID 根据 ID 查询单条历史记录
func (r *HistoryRepository) GetByID(id int64) (*HistoryRecord, error) {
	query := `
		SELECT
			id, code, name, entry_date, entry_price, exit_date, exit_price,
			holding_days, holding_duration, total_return, data_source, created_at
		FROM history
		WHERE id = ?;
	`
	row := r.db.QueryRow(query, id)
	return r.scanRow(row)
}

// Delete 根据 ID 删除历史记录
func (r *HistoryRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM history WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("delete history: %w", err)
	}
	return nil
}

// Count 返回历史记录总数量
func (r *HistoryRepository) Count() (int64, error) {
	var count int64
	row := r.db.QueryRow("SELECT COUNT(*) FROM history;")
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count history: %w", err)
	}
	return count, nil
}

// ClearAll 清空所有历史记录
func (r *HistoryRepository) ClearAll() error {
	_, err := r.db.Exec("DELETE FROM history;")
	if err != nil {
		return fmt.Errorf("clear all history: %w", err)
	}
	return nil
}

// scanRow 从 sql.Row 扫描一条 HistoryRecord 记录
func (r *HistoryRepository) scanRow(row *sql.Row) (*HistoryRecord, error) {
	var h HistoryRecord
	err := row.Scan(
		&h.ID, &h.Code, &h.Name, &h.EntryDate, &h.EntryPrice, &h.ExitDate, &h.ExitPrice,
		&h.HoldingDays, &h.HoldingDuration, &h.TotalReturn, &h.DataSource, &h.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan history: %w", err)
	}
	return &h, nil
}

// scanRows 从 sql.Rows 扫描多条 HistoryRecord 记录
func (r *HistoryRepository) scanRows(rows *sql.Rows) ([]HistoryRecord, error) {
	records := make([]HistoryRecord, 0)
	for rows.Next() {
		var h HistoryRecord
		if err := rows.Scan(
			&h.ID, &h.Code, &h.Name, &h.EntryDate, &h.EntryPrice, &h.ExitDate, &h.ExitPrice,
			&h.HoldingDays, &h.HoldingDuration, &h.TotalReturn, &h.DataSource, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan history rows: %w", err)
		}
		records = append(records, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history rows: %w", err)
	}
	return records, nil
}
