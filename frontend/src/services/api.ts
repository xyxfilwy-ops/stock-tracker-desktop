/**
 * Wails Bindings — 前端到 Go 后端的 API 封装
 *
 * 所有函数直接通过 window.go.<AppName>.<Method> 调用。
 * Go 端由 Wails 的 `runtime.EventsEmit` + `runtime.EventsOn` 提供进度回调。
 */

// ---------------------------------------------------------------------------
// TypeScript 类型定义
// ---------------------------------------------------------------------------

/** 当前持仓股票 */
export interface Stock {
  id: number;
  code: string;           // 如 "sh600519"
  name: string;           // 股票名称
  entryDate: string;      // 选入日期 "2026-06-19"
  entryTime: string;      // 选入时间 "14:30:00"
  entryPrice: number;     // 选入价（前复权，分）
  rawPrice: number;       // 选入日实际收盘价（不复权，分）
  adjustFactor: number;   // 复权因子（BP，1000=1.0）
  currentPrice: number;   // 当前价（分）
  prevClose: number;      // 昨收价（分）
  dailyChange: number;    // 今日涨跌幅（BP，148=1.48%）
  accChange: number;      // 累计涨跌幅（BP）
  dataSource: string;     // 当前数据源
  status: string;         // normal / suspended / ipo / ex-rights
  lastUpdate: string;     // 最后更新时间
}

/** 搜索返回的候选结果 */
export interface SearchResult {
  code: string;    // 如 "sh600519"
  name: string;    // 如 "贵州茅台"
  pinyin: string;  // 如 "GZMT"
  type: string;    // "stock" | "fund"
}

/** 已调出历史记录 */
export interface HistoryRecord {
  id: number;
  code: string;
  name: string;
  entryDate: string;       // 选入日期 YYYY-MM-DD
  entryTime: string;       // 选入时间 HH:MM:SS
  entryPrice: number;      // 前复权选入价（分）
  exitDate: string;        // 调出日期 YYYY-MM-DD
  exitTime: string;        // 调出时间 HH:MM:SS
  exitPrice: number;       // 调出价（分）
  holdingDays: number;     // 持股天数
  holdingDuration: string; // 持仓时间（如"1年3个月15天"）
  totalReturn: number;     // 区间收益（BP）
}

/** 批量刷新结果 */
export interface RefreshResult {
  updated: number;        // 成功更新数量
  failed: number;         // 失败数量
  total: number;          // 总数
  details: RefreshDetail[];
}

export interface RefreshDetail {
  id: number;
  code: string;
  name: string;
  error?: string;
}

/** 市场状态 */
export interface MarketStatus {
  status: string;         // "trading" | "closed" | "auction" | "lunch"
  message: string;
  isTrading: boolean;
}

// ---------------------------------------------------------------------------
// 全局 Wails 运行时声明（由 Wails 编译时注入）
// ---------------------------------------------------------------------------

declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetStocks(): Promise<Stock[]>;
          GetHistory(): Promise<HistoryRecord[]>;
          AddStock(code: string): Promise<Stock>;
          RemoveStock(id: number): Promise<HistoryRecord>;
          RefreshAll(): Promise<RefreshResult>;
          RefreshStock(id: number): Promise<Stock>;
          GetMarketStatus(): Promise<MarketStatus>;
          ClearHistory(): Promise<void>;
          SearchStocks(keyword: string): Promise<SearchResult[]>;
        };
      };
    };
    /** Wails 事件系统 — 用于接收进度回调 */
    runtime: {
      EventsOn(event: string, callback: (data: unknown) => void): () => void;
      EventsEmit(event: string, data: unknown): void;
    };
  }
}

// ---------------------------------------------------------------------------
// 统一错误包装
// ---------------------------------------------------------------------------

export class APIError extends Error {
  constructor(
    message: string,
    public readonly code?: string,
    public readonly cause?: unknown
  ) {
    super(message);
    this.name = 'APIError';
  }
}

function wrap<T>(promise: Promise<T>, operation: string): Promise<T> {
  return promise.catch((err: unknown) => {
    const msg = err instanceof Error ? err.message : String(err);
    throw new APIError(`${operation} 失败: ${msg}`, undefined, err);
  });
}

// ---------------------------------------------------------------------------
// 公开 API 函数
// ---------------------------------------------------------------------------

export function GetStocks(): Promise<Stock[]> {
  return wrap(window.go.main.App.GetStocks(), '获取持仓列表');
}

export function GetHistory(): Promise<HistoryRecord[]> {
  return wrap(window.go.main.App.GetHistory(), '获取历史记录');
}

export function AddStock(code: string): Promise<Stock> {
  return wrap(window.go.main.App.AddStock(code), '选入股票');
}

export function RemoveStock(id: number): Promise<HistoryRecord> {
  return wrap(window.go.main.App.RemoveStock(id), '调出股票');
}

export function RefreshAll(): Promise<RefreshResult> {
  return wrap(window.go.main.App.RefreshAll(), '刷新行情');
}

export function RefreshStock(id: number): Promise<Stock> {
  return wrap(window.go.main.App.RefreshStock(id), '刷新单只股票');
}

export function GetMarketStatus(): Promise<MarketStatus> {
  return wrap(window.go.main.App.GetMarketStatus(), '获取市场状态');
}

export function ClearHistory(): Promise<void> {
  return wrap(window.go.main.App.ClearHistory(), '清空历史记录');
}

export function SearchStocks(keyword: string): Promise<SearchResult[]> {
  return wrap(window.go.main.App.SearchStocks(keyword), '搜索股票/基金');
}

// ---------------------------------------------------------------------------
// 进度事件订阅（RefreshAll 过程中推送）
// ---------------------------------------------------------------------------

export interface RefreshProgress {
  current: number;
  total: number;
  code: string;
  success: boolean;
}

let unsubProgress: (() => void) | null = null;

/** 订阅刷新进度事件 */
export function onRefreshProgress(callback: (p: RefreshProgress) => void): () => void {
  // 先取消旧的订阅
  if (unsubProgress) {
    unsubProgress();
  }

  const off = window.runtime.EventsOn('refresh:progress', (data: unknown) => {
    const p = data as RefreshProgress;
    callback(p);
  });

  unsubProgress = off;
  return off;
}

/** 取消订阅 */
export function offRefreshProgress(): void {
  if (unsubProgress) {
    unsubProgress();
    unsubProgress = null;
  }
}

// ---------------------------------------------------------------------------
// 格式化工具（纯函数，供前端 UI 使用）
// ---------------------------------------------------------------------------

/** 分 → 元（显示 2 位小数） */
export function fmtPrice(cents: number): string {
  return (cents / 100).toFixed(2);
}

/** BP → %（如 148 → "+1.48%"） */
export function fmtChange(bp: number): string {
  const sign = bp >= 0 ? '+' : '';
  return `${sign}${(bp / 100).toFixed(2)}%`;
}

/** 涨跌方向（用于 CSS 类名） */
export function changeClass(bp: number): string {
  if (bp > 0) return 'positive';
  if (bp < 0) return 'negative';
  return 'neutral';
}

/** 状态中文映射 */
export function statusText(status: string): string {
  const map: Record<string, string> = {
    normal: '正常',
    suspended: '停牌',
    ipo: '上市首日',
    'ex-rights': '除权除息',
  };
  return map[status] || status;
}

export default {
  GetStocks,
  GetHistory,
  AddStock,
  RemoveStock,
  RefreshAll,
  RefreshStock,
  GetMarketStatus,
  onRefreshProgress,
  offRefreshProgress,
  ClearHistory,
  SearchStocks,
  fmtPrice,
  fmtChange,
  changeClass,
  statusText,
};
