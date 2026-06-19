package database

import (
	"database/sql"
	"fmt"
)

// StockRepository 提供 stocks 表的 CRUD 操作
type StockRepository struct {
	db *DB
}

// NewStockRepository 创建 StockRepository 实例
func NewStockRepository(db *DB) *StockRepository {
	return &StockRepository{db: db}
}

// Create 向 stocks 表插入一条新记录
func (r *StockRepository) Create(s *Stock) (*Stock, error) {
	query := `
		INSERT INTO stocks (
			code, name, entry_date, entry_time, entry_price, raw_price,
			adjust_factor, current_price, prev_close, daily_change, acc_change,
			data_source, status, last_update, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id;
	`
	var id int64
	err := r.db.QueryRow(query,
		s.Code, s.Name, s.EntryDate, s.EntryTime, s.EntryPrice, s.RawPrice,
		s.AdjustFactor, s.CurrentPrice, s.PrevClose, s.DailyChange, s.AccChange,
		s.DataSource, s.Status, s.LastUpdate, s.CreatedAt, s.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert stock: %w", err)
	}
	s.ID = id
	return s, nil
}

// GetAll 返回所有当前持仓记录，按创建时间升序排列
func (r *StockRepository) GetAll() ([]Stock, error) {
	query := `
		SELECT
			id, code, name, entry_date, entry_time, entry_price, raw_price,
			adjust_factor, current_price, prev_close, daily_change, acc_change,
			data_source, status, last_update, created_at, updated_at
		FROM stocks
		ORDER BY created_at ASC;
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query stocks: %w", err)
	}
	defer rows.Close()

	return r.scanRows(rows)
}

// GetByID 根据 ID 查询单条持仓记录
func (r *StockRepository) GetByID(id int64) (*Stock, error) {
	query := `
		SELECT
			id, code, name, entry_date, entry_time, entry_price, raw_price,
			adjust_factor, current_price, prev_close, daily_change, acc_change,
			data_source, status, last_update, created_at, updated_at
		FROM stocks
		WHERE id = ?;
	`
	row := r.db.QueryRow(query, id)
	return r.scanRow(row)
}

// GetByCode 根据股票代码查询单条持仓记录
func (r *StockRepository) GetByCode(code string) (*Stock, error) {
	query := `
		SELECT
			id, code, name, entry_date, entry_time, entry_price, raw_price,
			adjust_factor, current_price, prev_close, daily_change, acc_change,
			data_source, status, last_update, created_at, updated_at
		FROM stocks
		WHERE code = ?;
	`
	row := r.db.QueryRow(query, code)
	return r.scanRow(row)
}

// Update 更新持仓记录的所有可写字段（不包括 ID、created_at）
func (r *StockRepository) Update(s *Stock) error {
	query := `
		UPDATE stocks SET
			code = ?,
			name = ?,
			entry_date = ?,
			entry_time = ?,
			entry_price = ?,
			raw_price = ?,
			adjust_factor = ?,
			current_price = ?,
			prev_close = ?,
			daily_change = ?,
			acc_change = ?,
			data_source = ?,
			status = ?,
			last_update = ?,
			updated_at = ?
		WHERE id = ?;
	`
	_, err := r.db.Exec(query,
		s.Code, s.Name, s.EntryDate, s.EntryTime, s.EntryPrice, s.RawPrice,
		s.AdjustFactor, s.CurrentPrice, s.PrevClose, s.DailyChange, s.AccChange,
		s.DataSource, s.Status, s.LastUpdate, s.UpdatedAt, s.ID,
	)
	if err != nil {
		return fmt.Errorf("update stock: %w", err)
	}
	return nil
}

// Delete 根据 ID 删除持仓记录（daily_snapshots 会自动级联删除）
func (r *StockRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM stocks WHERE id = ?;", id)
	if err != nil {
		return fmt.Errorf("delete stock: %w", err)
	}
	return nil
}

// UpdatePrice 更新当前价格和昨收价（用于日常刷新）
func (r *StockRepository) UpdatePrice(id int64, currentPrice, prevClose int64, updatedAt string) error {
	query := `
		UPDATE stocks
		SET current_price = ?,
			prev_close = ?,
			updated_at = ?
		WHERE id = ?;
	`
	_, err := r.db.Exec(query, currentPrice, prevClose, updatedAt, id)
	if err != nil {
		return fmt.Errorf("update stock price: %w", err)
	}
	return nil
}

// UpdateAfterRefresh 更新刷新后的完整数据：现价、昨收、日涨跌幅、累计涨跌幅、数据源、状态、更新时间
func (r *StockRepository) UpdateAfterRefresh(id int64, currentPrice, prevClose, dailyChange, accChange int64, dataSource, status, lastUpdate, updatedAt string) error {
	query := `
		UPDATE stocks
		SET current_price = ?,
			prev_close = ?,
			daily_change = ?,
			acc_change = ?,
			data_source = ?,
			status = ?,
			last_update = ?,
			updated_at = ?
		WHERE id = ?;
	`
	_, err := r.db.Exec(query,
		currentPrice, prevClose, dailyChange, accChange,
		dataSource, status, lastUpdate, updatedAt, id,
	)
	if err != nil {
		return fmt.Errorf("update after refresh: %w", err)
	}
	return nil
}

// UpdateName 当 API 返回的名称与数据库不一致时自动更新名称
func (r *StockRepository) UpdateName(id int64, name string) error {
	query := `UPDATE stocks SET name = ? WHERE id = ?;`
	_, err := r.db.Exec(query, name, id)
	if err != nil {
		return fmt.Errorf("update stock name: %w", err)
	}
	return nil
}

// Count 返回当前持仓数量
func (r *StockRepository) Count() (int64, error) {
	var count int64
	row := r.db.QueryRow("SELECT COUNT(*) FROM stocks;")
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count stocks: %w", err)
	}
	return count, nil
}

// scanRow 从 sql.Row 扫描一条 Stock 记录
func (r *StockRepository) scanRow(row *sql.Row) (*Stock, error) {
	var s Stock
	err := row.Scan(
		&s.ID, &s.Code, &s.Name, &s.EntryDate, &s.EntryTime, &s.EntryPrice, &s.RawPrice,
		&s.AdjustFactor, &s.CurrentPrice, &s.PrevClose, &s.DailyChange, &s.AccChange,
		&s.DataSource, &s.Status, &s.LastUpdate, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan stock: %w", err)
	}
	return &s, nil
}

// scanRows 从 sql.Rows 扫描多条 Stock 记录
func (r *StockRepository) scanRows(rows *sql.Rows) ([]Stock, error) {
	stocks := make([]Stock, 0)
	for rows.Next() {
		var s Stock
		if err := rows.Scan(
			&s.ID, &s.Code, &s.Name, &s.EntryDate, &s.EntryTime, &s.EntryPrice, &s.RawPrice,
			&s.AdjustFactor, &s.CurrentPrice, &s.PrevClose, &s.DailyChange, &s.AccChange,
			&s.DataSource, &s.Status, &s.LastUpdate, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stock rows: %w", err)
		}
		stocks = append(stocks, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stock rows: %w", err)
	}
	return stocks, nil
}
