<script lang="ts">
  import type { Stock } from '../services/api';

  export let stocks: Stock[] = [];
  export let loading: boolean = false;
  export let onSelect: (stock: Stock) => void;

  function formatPrice(price: number): string {
    return (price / 100).toFixed(2);
  }

  function formatChange(bp: number): string {
    const pct = (bp / 100).toFixed(2);
    return bp >= 0 ? `+${pct}%` : `${pct}%`;
  }

  function getChangeColor(bp: number): string {
    if (bp > 0) return 'var(--positive, #dc2626)';
    if (bp < 0) return 'var(--negative, #16a34a)';
    return 'var(--text-secondary, #6b7280)';
  }

  function isErrorRow(stock: Stock): boolean {
    return stock.status === 'normal' && stock.lastUpdate === undefined && stock.dataSource === undefined;
  }

  function handleRowClick(stock: Stock) {
    if (stock.status === 'suspended') return;
    onSelect(stock);
  }
</script>

<div class="table-wrapper">
  <table class="stock-table">
    <thead>
      <tr class="header-row">
        <th class="cell" style="width: 80px;">代码</th>
        <th class="cell" style="width: 100px;">名称</th>
        <th class="cell" style="width: 80px; text-align: right;">选入价</th>
        <th class="cell" style="width: 80px; text-align: right;">现价</th>
        <th class="cell" style="width: 80px; text-align: right;">日涨跌</th>
        <th class="cell" style="width: 80px; text-align: right;">累计涨跌</th>
        <th class="cell" style="width: 40px;"></th>
      </tr>
    </thead>
    <tbody>
      {#each stocks as stock (stock.id)}
        <tr
          class="data-row"
          class:suspended={stock.status === 'suspended'}
          class:error={isErrorRow(stock)}
          on:click={() => handleRowClick(stock)}
          class:clickable={stock.status !== 'suspended'}
        >
          <td class="cell code">{stock.code}</td>
          <td class="cell name">{stock.name}</td>
          <td class="cell" style="text-align: right;">{formatPrice(stock.entryPrice)}</td>
          <td class="cell" style="text-align: right;">
            {#if stock.status === 'suspended'}
              <span class="muted">—</span>
            {:else}
              {formatPrice(stock.currentPrice)}
            {/if}
          </td>
          <td class="cell" style="text-align: right;">
            {#if stock.status === 'suspended'}
              <span class="muted">—</span>
            {:else}
              <span class="badge" style="color: {getChangeColor(stock.dailyChange)};">
                {formatChange(stock.dailyChange)}
              </span>
            {/if}
          </td>
          <td class="cell" style="text-align: right;">
            {#if stock.status === 'suspended'}
              <span class="muted">—</span>
            {:else}
              <span class="badge" style="color: {getChangeColor(stock.accChange)};">
                {formatChange(stock.accChange)}
              </span>
            {/if}
          </td>
          <td class="cell indicator">
            {#if stock.status === 'suspended'}
              <span class="suspended-badge" title="停牌">⏸</span>
            {:else if isErrorRow(stock)}
              <span class="error-icon" title="获取失败">⚠️</span>
            {:else if loading}
              <span class="pulse-dot"></span>
            {/if}
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<style>
  .table-wrapper {
    width: 100%;
    overflow-x: auto;
  }

  .stock-table {
    width: 100%;
    border-collapse: collapse;
    font-family: var(--font-mono, 'JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', Consolas, monospace);
    font-size: 14px;
  }

  .header-row {
    height: 36px;
    background: transparent;
    border-bottom: 1px solid var(--border, #e2e4e8);
  }

  .header-row th {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary, #6b7280);
    text-align: left;
    padding: 8px 12px;
    white-space: nowrap;
  }

  .data-row {
    height: 44px;
    border-bottom: 1px solid var(--border, #e2e4e8);
    transition: background-color 150ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .data-row:hover {
    background-color: var(--hover-bg, #f3f4f6);
  }

  .data-row.clickable {
    cursor: pointer;
  }

  .data-row.suspended {
    color: var(--text-muted, #9ca3af);
  }

  .data-row.error .code,
  .data-row.error .name {
    color: var(--text-muted, #9ca3af);
  }

  .cell {
    padding: 8px 12px;
    white-space: nowrap;
    vertical-align: middle;
  }

  .code {
    font-family: var(--font-mono, 'JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', Consolas, monospace);
    font-size: 13px;
  }

  .name {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-weight: 500;
  }

  .badge {
    font-family: var(--font-mono, 'JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', Consolas, monospace);
    font-weight: 600;
    font-size: 13px;
  }

  .muted {
    color: var(--text-muted, #9ca3af);
  }

  .pulse-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--primary, #2563eb);
    animation: pulse 1.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; transform: scale(1); }
    50% { opacity: 0.4; transform: scale(0.8); }
  }

  .error-icon {
    color: var(--warning, #d97706);
    cursor: help;
  }

  .suspended-badge {
    color: var(--text-muted, #9ca3af);
    font-size: 12px;
  }

  .indicator {
    text-align: center;
    padding: 8px 4px;
  }
</style>
