<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import './styles/design.css';
  import {
    GetStocks,
    GetHistory,
    AddStock,
    RemoveStock,
    RefreshAll,
    GetMarketStatus,
    onRefreshProgress,
    offRefreshProgress,
    type Stock,
    type HistoryRecord,
    type RefreshResult,
    type MarketStatus,
    fmtPrice,
    fmtChange,
    changeClass,
  } from './services/api';

  import Header from './components/Header.svelte';
  import StockTable from './components/StockTable.svelte';
  import HistoryTable from './components/HistoryTable.svelte';
  import AddDialog from './components/AddDialog.svelte';
  import ConfirmDialog from './components/ConfirmDialog.svelte';
  import EmptyState from './components/EmptyState.svelte';
  import ErrorBanner from './components/ErrorBanner.svelte';

  // ---------------------------------------------------------------------------
  // 全局状态
  // ---------------------------------------------------------------------------
  let activeTab: 'holdings' | 'history' = 'holdings';
  let stocks: Stock[] = [];
  let historyRecords: HistoryRecord[] = [];
  let loading = false;
  let refreshProgress = { current: 0, total: 0 };
  let toast = { message: '', type: 'success' as 'success' | 'error' | 'info', visible: false };
  let banner = { message: '', visible: false };
  let showAddDialog = false;
  let showConfirmDialog = false;
  let selectedStock: Stock | null = null;
  let toastTimer: ReturnType<typeof setTimeout> | null = null;
  let autoRefreshTimer: ReturnType<typeof setInterval> | null = null;
  let marketStatus: MarketStatus | null = null;
  let addDialogError = '';
  let addDialogLoading = false;

  // ---------------------------------------------------------------------------
  // 派生状态
  // ---------------------------------------------------------------------------
  $: avgAccChange = stocks.length > 0
    ? Math.round(stocks.reduce((sum, s) => sum + (s.accChange || 0), 0) / stocks.length)
    : 0;

  $: showEmptyHoldings = activeTab === 'holdings' && stocks.length === 0 && !loading;
  $: showEmptyHistory = activeTab === 'history' && historyRecords.length === 0;

  $: refreshPercent = refreshProgress.total > 0
    ? (refreshProgress.current / refreshProgress.total) * 100
    : 0;

  // ---------------------------------------------------------------------------
  // 生命周期
  // ---------------------------------------------------------------------------
  let initError = '';

  onMount(() => {
    // 检查 Wails 绑定是否就绪
    if (typeof window === 'undefined' || !window.go || !window.go.main || !window.go.main.App) {
      initError = '应用初始化失败：后端绑定未就绪，请重启程序';
      console.error('window.go not available:', typeof window !== 'undefined' ? window.go : 'window undefined');
      return;
    }
    loadStocks();
    loadHistory();
    checkMarketStatus();
  });

  onDestroy(() => {
    if (toastTimer) clearTimeout(toastTimer);
    if (autoRefreshTimer) clearInterval(autoRefreshTimer);
    offRefreshProgress();
  });

  // ---------------------------------------------------------------------------
  // 数据加载
  // ---------------------------------------------------------------------------
  async function loadStocks() {
    try {
      const result = await GetStocks();
      stocks = result || [];
    } catch (err) {
      console.error('loadStocks failed:', err);
      stocks = [];
      showBanner('数据加载失败，请稍后重试');
    }
  }

  async function loadHistory() {
    try {
      const result = await GetHistory();
      historyRecords = result || [];
    } catch (err) {
      console.error('loadHistory failed:', err);
      historyRecords = [];
    }
  }

  async function checkMarketStatus() {
    try {
      marketStatus = await GetMarketStatus();
      if (marketStatus?.isTrading && !autoRefreshTimer) {
        autoRefreshTimer = setInterval(() => {
          if (!loading) handleRefresh();
        }, 60000);
      } else if (!marketStatus?.isTrading && autoRefreshTimer) {
        clearInterval(autoRefreshTimer);
        autoRefreshTimer = null;
      }
    } catch (err) {
      console.error('checkMarketStatus failed:', err);
    }
  }

  // ---------------------------------------------------------------------------
  // 刷新行情
  // ---------------------------------------------------------------------------
  async function handleRefresh() {
    if (loading) return;
    loading = true;
    refreshProgress = { current: 0, total: stocks.length };
    hideBanner();

    onRefreshProgress((p: { current: number; total: number }) => {
      refreshProgress = { current: p.current, total: p.total };
    });

    try {
      const result: RefreshResult = await RefreshAll();
      await loadStocks();

      if (result.failed === 0) {
        showToast(`已更新 ${result.updated}/${result.total} 只股票`, 'success');
      } else if (result.updated > 0) {
        showToast(`已更新 ${result.updated}/${result.total} 只，${result.failed} 只获取失败`, 'error');
      } else {
        showBanner('行情源暂时不可用，数据为上次缓存');
      }
    } catch (err) {
      console.error('handleRefresh failed:', err);
      showBanner('行情源暂时不可用，数据为上次缓存');
    } finally {
      loading = false;
      refreshProgress = { current: 0, total: 0 };
      offRefreshProgress();
    }
  }

  // ---------------------------------------------------------------------------
  // 选入 / 调出
  // ---------------------------------------------------------------------------
  async function handleAddStock(code: string) {
    addDialogLoading = true;
    addDialogError = '';
    try {
      const newStock = await AddStock(code);
      stocks = [...stocks, newStock];
      showToast(`已选入 ${newStock.name} (${newStock.code})`, 'success');
      showAddDialog = false;
      addDialogLoading = false;
      addDialogError = '';
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      addDialogError = msg;
      addDialogLoading = false;
      // 弹窗不关闭，显示错误
    }
  }

  async function handleRemoveStock(id: number) {
    try {
      const record = await RemoveStock(id);
      stocks = stocks.filter((s) => s.id !== id);
      historyRecords = [record, ...historyRecords];
      showToast(`已调出 ${record.name}，区间收益 ${fmtChange(record.totalReturn)}`, 'success');
      showConfirmDialog = false;
      selectedStock = null;
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      showToast(msg, 'error');
    }
  }

  // ---------------------------------------------------------------------------
  // 对话框控制
  // ---------------------------------------------------------------------------
  function openAddDialog() {
    showAddDialog = true;
    addDialogError = '';
    addDialogLoading = false;
  }

  function closeAddDialog() {
    showAddDialog = false;
    addDialogError = '';
    addDialogLoading = false;
  }

  function openConfirmDialog(stock: Stock) {
    selectedStock = stock;
    showConfirmDialog = true;
  }

  function closeConfirmDialog() {
    showConfirmDialog = false;
    selectedStock = null;
  }

  // ---------------------------------------------------------------------------
  // 提示展示
  // ---------------------------------------------------------------------------
  function showToast(message: string, type: 'success' | 'error' | 'info' = 'success') {
    if (toastTimer) clearTimeout(toastTimer);
    toast = { message, type, visible: true };
    toastTimer = setTimeout(() => {
      toast = { ...toast, visible: false };
    }, 5000);
  }

  function showBanner(message: string) {
    banner = { message, visible: true };
  }

  function hideBanner() {
    banner = { ...banner, visible: false };
  }
</script>

<svelte:head>
  <title>StockTracker</title>
</svelte:head>

<!-- 顶部刷新进度条 -->
{#if loading && refreshProgress.total > 0}
  <div class="refresh-progress-bar" style="width: {refreshPercent}%"></div>
{/if}

<!-- Toast 提示 -->
{#if toast.visible}
  <div class="toast toast-{toast.type}">
    {toast.message}
  </div>
{/if}

<div id="app">
  <!-- 顶栏 — 只保留标题，不加按钮 -->
  <Header />

  <!-- 初始化错误提示 -->
  {#if initError}
    <div class="init-error">
      <span class="init-error-icon">⚠</span>
      <span class="init-error-text">{initError}</span>
    </div>
  {/if}

  <!-- Tab 导航 -->
  <nav class="tab-nav">
    <button
      class="tab-btn"
      class:active={activeTab === 'holdings'}
      on:click={() => (activeTab = 'holdings')}
      type="button"
    >
      当前持仓
    </button>
    <button
      class="tab-btn"
      class:active={activeTab === 'history'}
      on:click={() => (activeTab = 'history')}
      type="button"
    >
      历史记录
    </button>
  </nav>

  <!-- 错误 Banner -->
  {#if banner.visible}
    <ErrorBanner message={banner.message} onDismiss={hideBanner} />
  {/if}

  <!-- 内容区 -->
  <main class="content">
    {#if activeTab === 'holdings'}
      <div class="panel">
        <div class="panel-header">
          <div class="panel-header-left">
            <span class="panel-title">持仓概览</span>
            {#if stocks.length > 0}
              <span class="panel-count">共 {stocks.length} 只</span>
            {/if}
          </div>
          <div class="panel-header-right">
            {#if stocks.length > 0}
              <div class="summary-text">
                合计:
                <span class="change {changeClass(avgAccChange)}">
                  {fmtChange(avgAccChange)}
                </span>
              </div>
            {/if}
            <button class="btn btn-primary" on:click={openAddDialog} type="button">
              + 选入
            </button>
          </div>
        </div>
        {#if stocks.length > 0}
          <StockTable {stocks} {loading} onSelect={openConfirmDialog} />
        {:else}
          <EmptyState
            title="还没有跟踪任何股票"
            description="点击「选入股票」添加你关注的股票，开始跟踪完整生命周期涨跌幅。"
            hint="支持 sh/sz 开头或纯数字代码，例如：600519 或 sh600519"
            onAdd={openAddDialog}
          />
        {/if}
      </div>

    {:else if activeTab === 'history'}
      <div class="panel">
        {#if historyRecords.length > 0}
          <div class="panel-header">
            <span class="panel-title">历史记录</span>
            <span class="panel-meta">共 {historyRecords.length} 条</span>
          </div>
          <HistoryTable history={historyRecords} />
        {:else}
          <EmptyState
            title="暂无历史记录"
            description="调出股票后将自动移入此处，你可以回顾每一次选入→调出的完整表现。"
            hint=""
          />
        {/if}
      </div>
    {/if}
  </main>
</div>

<!-- 选入对话框 -->
<AddDialog
  visible={showAddDialog}
  onConfirm={handleAddStock}
  onCancel={closeAddDialog}
  error={addDialogError}
  loading={addDialogLoading}
/>

<!-- 调出确认对话框 -->
<ConfirmDialog
  visible={showConfirmDialog}
  stock={selectedStock}
  onConfirm={() => selectedStock && handleRemoveStock(selectedStock.id)}
  onCancel={closeConfirmDialog}
/>

<!-- 右下角浮动刷新按钮 -->
<button
  class="fab"
  class:fab-spinning={loading}
  on:click={handleRefresh}
  disabled={loading}
  title="刷新行情"
  type="button"
>
  <span class="fab-icon" class:spinning={loading}>⟳</span>
</button>

<style>
  .panel-header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .panel-count {
    font-size: 13px;
    color: var(--ink-400, #9ca3af);
    font-weight: 400;
  }

  .panel-header-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 34px;
    padding: 0 14px;
    border-radius: 8px;
    font-family: var(--font-body, 'Inter', sans-serif);
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    border: none;
    outline: none;
    transition: all 120ms cubic-bezier(0.4, 0, 0.2, 1);
    white-space: nowrap;
    gap: 6px;
  }

  .btn-primary {
    background: var(--primary, #1e293b);
    color: #fff;
  }
  .btn-primary:hover {
    background: var(--primary-hover, #0f172a);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md, 0 4px 12px rgba(0,0,0,0.04));
  }

  .summary-text {
    font-size: 14px;
    color: var(--ink-500);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .fab {
    position: fixed;
    bottom: 32px;
    right: 32px;
    width: 56px;
    height: 56px;
    border-radius: 999px;
    background: var(--primary);
    color: #fff;
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    box-shadow: var(--shadow-lg);
    transition: all 200ms cubic-bezier(0.4, 0, 0.2, 1);
    z-index: 100;
  }

  .fab:hover:not(:disabled) {
    transform: translateY(-2px) scale(1.05);
    box-shadow: var(--shadow-xl);
  }

  .fab:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none !important;
  }

  .fab-icon {
    display: inline-block;
    transition: transform 0.3s ease;
  }

  .fab-icon.spinning {
    animation: spin 1s linear infinite;
  }

  .init-error {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: var(--error-bg, #fef2f2);
    border: 1px solid rgba(185, 28, 28, 0.15);
    border-radius: var(--radius-md, 8px);
    color: var(--error-text, #b91c1c);
    font-size: 13px;
    font-weight: 500;
  }

  .init-error-icon {
    font-size: 16px;
  }

  .init-error-text {
    line-height: 1.5;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
