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

  /* 子组件（由 Worker_Frontend_Components 提供） */
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
  let activeTab: 'holdings' | 'history' | 'manage' = 'holdings';
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

  // ---------------------------------------------------------------------------
  // 派生状态
  // ---------------------------------------------------------------------------

  /** 持仓平均累计涨跌幅（简单平均） */
  $: avgAccChange = stocks.length > 0
    ? Math.round(stocks.reduce((sum, s) => sum + (s.accChange || 0), 0) / stocks.length)
    : 0;

  /** 是否显示空状态 */
  $: showEmpty = activeTab === 'holdings' && stocks.length === 0 && !loading;

  /** 刷新进度百分比 */
  $: refreshPercent = refreshProgress.total > 0
    ? (refreshProgress.current / refreshProgress.total) * 100
    : 0;

  // ---------------------------------------------------------------------------
  // 生命周期
  // ---------------------------------------------------------------------------

  onMount(() => {
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
      stocks = await GetStocks();
    } catch (err) {
      console.error('loadStocks failed:', err);
      showBanner('数据加载失败，请稍后重试');
    }
  }

  async function loadHistory() {
    try {
      historyRecords = await GetHistory();
    } catch (err) {
      console.error('loadHistory failed:', err);
    }
  }

  async function checkMarketStatus() {
    try {
      marketStatus = await GetMarketStatus();
      // 交易时段内启动自动刷新（60秒）
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

    // 订阅进度事件
    onRefreshProgress((p) => {
      refreshProgress = { current: p.current, total: p.total };
    });

    try {
      const result: RefreshResult = await RefreshAll();
      await loadStocks(); // 刷新完成后重新拉取最新数据

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
    try {
      const newStock = await AddStock(code);
      stocks = [...stocks, newStock];
      showToast(`已选入 ${newStock.name} (${newStock.code})`, 'success');
      showAddDialog = false;
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      showToast(msg, 'error');
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
  }

  function closeAddDialog() {
    showAddDialog = false;
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

  // ---------------------------------------------------------------------------
  // 工具：按 Tab 切换页面名称
  // ---------------------------------------------------------------------------
  function tabLabel(tab: typeof activeTab): string {
    const map = { holdings: '当前持仓', history: '历史记录', manage: '管理操作' };
    return map[tab];
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
  <!-- 顶栏 -->
  <Header {loading} onRefresh={handleRefresh} />

  <!-- Tab 导航 -->
  <nav class="tab-nav">
    <button
      class="tab-btn"
      class:active={activeTab === 'holdings'}
      on:click={() => (activeTab = 'holdings')}
    >
      当前持仓
    </button>
    <button
      class="tab-btn"
      class:active={activeTab === 'history'}
      on:click={() => (activeTab = 'history')}
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
        <!-- 合计行 + 选入按钮 -->
        <div class="panel-header">
          <button class="btn-primary" on:click={openAddDialog}>
            + 选入股票
          </button>
          {#if stocks.length > 0}
            <div class="summary-text">
              合计累计:
              <span class="change {changeClass(avgAccChange)}">
                {fmtChange(avgAccChange)}
              </span>
            </div>
          {/if}
        </div>

        {#if showEmpty}
          <EmptyState
            title="还没有跟踪任何股票"
            description="点击「+ 选入股票」开始跟踪"
            hint="支持 sh/sz 开头或纯数字代码，例如：600519 或 sh600519"
          />
        {:else}
          <StockTable {stocks} {loading} onSelect={openConfirmDialog} />
        {/if}
      </div>

    {:else if activeTab === 'history'}
      <div class="panel">
        <div class="panel-header">
          <span class="panel-title">历史记录</span>
          <span class="text-muted" style="font-size: 13px;">
            共 {historyRecords.length} 条
          </span>
        </div>
        {#if historyRecords.length === 0}
          <EmptyState
            title="暂无历史记录"
            description="调出股票后将自动移入此处"
            hint=""
          />
        {:else}
          <HistoryTable {historyRecords} />
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
  />

<!-- 调出确认对话框 -->
  <ConfirmDialog
    visible={showConfirmDialog}
    stock={selectedStock}
    onConfirm={() => selectedStock && handleRemoveStock(selectedStock.id)}
    onCancel={closeConfirmDialog}
  />

<style>
  /* 本组件仅保留布局相关样式，颜色/字体统一走 design.css */
  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .summary-text {
    font-size: 14px;
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .text-muted {
    color: var(--text-muted);
  }
</style>
