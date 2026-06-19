# StockTracker 桌面应用 — Loop 1 开发计划（Goal × Loop Engineering）

## Goal（目标定义）

> **做一个本地运行的桌面股票跟踪工具，记录每只股票从选入到调出的完整生命周期涨跌幅。**

本次聚焦 **Loop 1：MVP 核心功能**，验收标准 AC1~AC9 全部通过。

## 技术栈

- **后端**: Go 1.22+ + Wails v2
- **前端**: TypeScript + Svelte + Vite
- **数据库**: SQLite (modernc.org/sqlite, WAL模式)
- **数据源**: 腾讯财经（主源）→ 新浪（备用1）→ 东方财富（备用2）

## 项目目录

`E:\kimi项目\股票基金跟踪\stock-tracker-desktop`

## Loop 1 阶段划分

### Stage 1: 项目骨架与配置
**Goal**: 建立可编译的 Wails + Go + Svelte 项目基础
**Agent**: 主Agent（亲自完成）
**输出**:
- `go.mod`, `wails.json`, `main.go`, `app.go`
- `frontend/package.json`, `vite.config.ts`, `tsconfig.json`, `svelte.config.js`, `index.html`
- `frontend/src/main.ts`

### Stage 2: 后端核心（并行）
**Goal**: 完成 Go 后端所有业务逻辑
**并行 Workers**:

**Worker_DB** — 数据库层
- 文件: `database/db.go`, `database/models.go`, `database/migrations.go`, `database/stock_repo.go`, `database/history_repo.go`
- 要求: 严格遵循 PRD 表结构（整数分/BP 存储、WAL模式、索引、级联删除）

**Worker_Provider** — 数据源适配层
- 文件: `providers/provider.go`, `providers/tencent.go`, `providers/sina.go`, `providers/eastmoney.go`
- 要求: 统一 Quote 结构、三源容灾、熔断器逻辑、数据归一化、同一源锁定、东方财富前复权K线接口

**Worker_Service** — 业务逻辑层
- 文件: `services/stock_service.go`, `services/market_service.go`, `services/refresh_manager.go`, `services/calculator.go`, `services/trading_calendar.go`, `config/config.go`
- 要求: 交易时段判断、批量刷新（semaphore+errgroup）、涨跌幅计算（前复权）、选入/调出流程

### Stage 3: 前端核心（并行）
**Goal**: 完成 Svelte 前端所有页面与交互
**并行 Workers**:

**Worker_Frontend_Core** — 前端骨架与API
- 文件: `frontend/src/App.svelte`, `frontend/src/services/api.ts`, `frontend/src/styles/design.css`
- 要求: 三个Tab页面（持仓/历史/管理）、Wails bindings、DESIGN.md调色盘与排版

**Worker_Frontend_Components** — 组件层
- 文件: `frontend/src/components/Header.svelte`, `StockTable.svelte`, `HistoryTable.svelte`, `AddDialog.svelte`, `ConfirmDialog.svelte`, `EmptyState.svelte`, `RefreshIndicator.svelte`, `ErrorBanner.svelte`
- 要求: 严格遵循 PRD 组件规范（颜色、间距、动效、反模式）

### Stage 4: 整合验证
**Goal**: 确保代码完整、目录结构正确、README完备
**Agent**: 主Agent
**任务**:
1. 检查所有文件是否生成
2. 验证前后端接口一致性
3. 编写 README.md（编译/运行指南）
4. 输出 Loop 1 完成报告

## 依赖关系

```
Stage 1 → Stage 2 + Stage 3（并行）
Stage 2 + Stage 3 → Stage 4
```

## 关键规范（所有 Worker 必须遵循）

1. **价格存储**: 整数分（172500 = 1725.00元）
2. **涨跌幅存储**: 整数 BP（148 = 1.48%）
3. **前复权计算**: `entry_price_adjusted = entry_price * adjust_factor / 1000`
4. **AC1~AC9**: 每个验收标准必须能在代码中对应找到实现
5. **红涨绿跌**: A股惯例，positive=红色，negative=绿色
6. **错误分层**: cell-level → toast-level → banner-level → offline-level

## PRD 参考文件

`E:\kimi项目\股票基金跟踪\🎯 StockTracker — 本地股票跟踪桌面工具 · ....md`

所有 Worker 在执行前必须阅读该 PRD 文件。
