# StockTracker — 本地股票跟踪桌面工具

> **Loop 1 MVP 版本** | Go + Wails + Svelte + SQLite | 三源容灾

## 项目简介

StockTracker 是一个本地运行的桌面股票跟踪工具，记录每只股票从选入到调出的完整生命周期涨跌幅。

**核心功能（Loop 1 MVP）**：
- 📊 **当前持仓**：展示正在跟踪的股票列表，含每日涨跌幅和累计涨跌幅
- 📜 **历史记录**：已调出股票的完整区间表现
- ⚙️ **管理操作**：选入股票、调出股票、刷新行情

**技术特性**：
- ✅ 纯本地运行，数据不出本机（SQLite 本地存储）
- ✅ 三源容灾：腾讯财经 → 新浪 → 东方财富
- ✅ 前复权计算，真实反映持有期收益
- ✅ A股交易时段感知
- ✅ 红涨绿跌，符合 A 股惯例

## 目录结构

```
stock-tracker-desktop/
├── main.go                   # 程序入口
├── app.go                    # Wails 应用配置 + 前端绑定 API
├── wails.json                # Wails 项目配置
├── go.mod / go.sum           # Go 依赖
├── config/                   # 配置管理
│   └── config.go
├── database/                 # SQLite 数据库层
│   ├── db.go                 # 连接初始化（WAL 模式）
│   ├── models.go             # 数据模型 + 计算工具
│   ├── migrations.go         # 表结构迁移
│   ├── stock_repo.go         # 持仓表 Repository
│   └── history_repo.go       # 历史表 Repository
├── services/                 # 业务逻辑层
│   ├── stock_service.go      # 选入/调出/查询
│   ├── market_service.go     # 三源容灾调度 + 熔断器
│   ├── refresh_manager.go    # 批量刷新（semaphore + 并发）
│   ├── calculator.go         # 涨跌幅计算（纯函数）
│   └── trading_calendar.go   # A股交易时段判断
├── providers/                # 数据源适配器
│   ├── provider.go           # 统一接口 + 归一化层
│   ├── tencent.go            # 腾讯财经（主源）
│   ├── sina.go              # 新浪财经（备用1）
│   └── eastmoney.go          # 东方财富（备用2 + 前复权K线）
├── frontend/                 # Svelte + TypeScript 前端
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── src/
│       ├── main.ts
│       ├── App.svelte
│       ├── services/api.ts   # Wails bindings + 类型
│       ├── styles/design.css # OpenDesign 设计系统
│       └── components/
│           ├── Header.svelte
│           ├── StockTable.svelte
│           ├── HistoryTable.svelte
│           ├── AddDialog.svelte
│           ├── ConfirmDialog.svelte
│           ├── EmptyState.svelte
│           ├── RefreshIndicator.svelte
│           └── ErrorBanner.svelte
└── README.md
```

## 环境要求

- **Go 1.22+**
- **Node.js 18+** + **npm**
- **Wails CLI v2.9+**

### 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

确保 `~/go/bin` 或 `%USERPROFILE%\go\bin` 在 PATH 中。

## 构建与运行

### 1. 克隆项目

```bash
cd stock-tracker-desktop
```

### 2. 安装前端依赖

```bash
cd frontend
npm install
cd ..
```

### 3. 整理 Go 依赖

```bash
go mod tidy
```

### 4. 开发模式（热重载）

```bash
wails dev
```

### 5. 构建生产版本

```bash
# Windows
wails build -platform windows

# macOS
wails build -platform darwin

# Linux
wails build -platform linux
```

构建产物位于 `build/bin/` 目录。

## 使用说明

### 首次启动

1. 运行程序后，自动在用户数据目录创建 SQLite 数据库：
   - Windows: `%LOCALAPPDATA%\StockTracker\`
   - macOS: `~/Library/Application Support/StockTracker/`
   - Linux: `~/.local/share/StockTracker/`

2. 点击 **"+ 选入股票"** 按钮，输入股票代码（如 `600519` 或 `sh600519`）

3. 系统自动识别市场、获取行情、计算前复权价格并记录

### 日常操作

| 操作 | 说明 |
|------|------|
| **选入股票** | 点击 "+ 选入股票" → 输入代码 → 确认 |
| **刷新行情** | 点击顶栏 "⟳ 刷新" 按钮，或等待交易时段自动刷新 |
| **调出股票** | 在持仓列表中点击目标股票行 → 确认调出 |
| **查看历史** | 切换至 "📜 历史记录" Tab |

### 数据展示

- **价格**：保留 2 位小数（如 1725.00 元）
- **涨跌幅**：百分比形式（如 +1.48%）
- **红涨绿跌**：符合 A 股惯例
- **停牌标记**：⏸ 停牌，价格显示 "—"
- **错误提示**：⚠️ 抓取失败，数据保持旧值

## 数据源说明

| 优先级 | 源 | 接口 | 特点 |
|--------|-----|------|------|
| 1 | 腾讯财经 | `qt.gtimg.cn` | 主源，最稳定，GBK 编码 |
| 2 | 新浪财经 | `hq.sinajs.cn` | 备用1，需 Referer，GBK 编码 |
| 3 | 东方财富 | `push2.eastmoney.com` | 备用2，JSON，含前复权 K 线 |

- **同一源锁定**：每次刷新优先使用上次成功的数据源，保证昨收价一致性
- **熔断器**：连续 5 次失败率 ≥60% 时自动熔断 30 秒
- **健康探测**：后台每 60 秒检测各源可用性

## 设计系统

遵循 [OpenDesign](https://open-design.ai) 九段式规范：

- **Color**：OKLch 色彩空间，语义化命名（红涨绿跌）
- **Typography**：Inter + PingFang SC + Microsoft YaHei
- **Spacing**：8px 基础单位
- **Layout**：单列布局，最大宽度 720px，响应式 480px/720px 断点
- **Motion**：cubic-bezier(0.4, 0, 0.2, 1)，禁止弹性 easing
- **Anti-patterns**：禁止紫色渐变、emoji 涨跌、斑马纹、hover 文字放大

## 开发路线图（Goal × Loop）

| Loop | 目标 | 内容 | 状态 |
|------|------|------|------|
| **Loop 1** | 能用 | 选入/查看/调出/刷新，数据持久化，三源容灾 | ✅ 已完成 |
| **Loop 2** | 好用 | UI 美化、错误提示优化、自动刷新、键盘快捷键 | 待迭代 |
| **Loop 3** | 更丰富 | 每日快照历史、简易趋势图、数据导出 CSV | 待迭代 |

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 后端 | Go 1.22+ | 编译快、单文件分发、并发强 |
| 桌面框架 | Wails v2 | Go 原生桌面，内嵌 Webview |
| 前端 | TypeScript + Svelte 4 + Vite | 编译型框架、零运行时、响应式 |
| 数据库 | SQLite (modernc.org/sqlite) | 纯 Go 实现，零配置，WAL 模式 |
| 数据源 | 腾讯/新浪/东方财富 | 免费，三源容灾 |

## 许可

MIT License — 个人投资者免费使用。

> ⚠️ **免责声明**：本工具仅用于个人投资跟踪，不构成任何投资建议。股市有风险，投资需谨慎。
