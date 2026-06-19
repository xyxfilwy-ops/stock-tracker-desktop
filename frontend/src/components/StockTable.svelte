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
    if (bp > 0) return 'var(--positive, #b9403a)';
    if (bp < 0) return 'var(--negative, #2e8b57)';
    return 'var(--neutral, #6b7280)';
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
        <th class="cell">代码</th>
        <th class="cell">名称</th>
        <th class="cell" style="text-align: right;">选入价</th>
        <th class="cell" style="text-align: right;">现价</th>
        <th class="cell" style="text-align: right;">日涨跌</th>
        <th class="cell" style="text-align: right;">累计涨跌</th>
        <th class="cell" style="width: 40px;"></th>
      </tr>
    </thead>
    <tbody>
      {#each stocks as stock (stock.id)}
        <tr
          class="data-row"
          class:suspended={stock.status === 'suspended'}
          class:error={isErrorRow(stock)}
          class:clickable={stock.status !== 'suspended'}
          on:click={() => handleRowClick(stock)}
        >
          <td class="cell code">{stock.code}</td>
          <td class="cell name">{stock.name}</td>
          <td class="cell" style="text-align: right;">
            <span class="price">{formatPrice(stock.entryPrice)}</span>
          </td>
          <td class="cell" style="text-align: right;">
            {#if stock.status === 'suspended'}
              <span class="muted">—</span>
            {:else}
              <span class="price">{formatPrice(stock.currentPrice)}</span>
            {/if}
          </td>
          <td class="cell" style="text-align: right;">
            {#if stock.status === 'suspended'}
              <span class="muted">—</span>
            {:else}
              <span class="badge badge-up" style="color: {getChangeColor(stock.dailyChange)}; background: {stock.dailyChange > 0 ? 'var(--positive-bg)' : stock.dailyChange < 0 ? 'var(--negative-bg)' : 'rgba(107,114,128,0.06)'};">
                {formatChange(stock.dailyChange)}
              </span>
            {/if}
          </td>
          <td class="cell" style="text-align: right;">
            {#if stock.status === 'suspended'}
              <span class="muted">—</span>
            {:else}
              <span class="badge badge-up" style="color: {getChangeColor(stock.accChange)}; background: {stock.accChange > 0 ? 'var(--positive-bg)' : stock.accChange < 0 ? 'var(--negative-bg)' : 'rgba(107,114,128,0.06)'};">
                {formatChange(stock.accChange)}
              </span>
            {/if}
          </td>
          <td class="cell indicator">
            {#if stock.status === 'suspended'}
              <span class="suspended-badge">停牌</span>
            {:else if isErrorRow(stock)}
              <span class="error-icon" title="获取失败">!</span>
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
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none;
  }
  .table-wrapper::-webkit-scrollbar {
    display: none;
  }

  .stock-table {
    width: 100%;
    border-collapse: separate;
    border-spacing: 0;
    font-size: 14px;
  }

  .header-row th {
    font-family: var(--font-body, 'Inter', sans-serif);
    font-size: 11px;
    font-weight: 600;
    color: var(--ink-400, #9ca3af);
    text-align: left;
    padding: 12px 16px;
    white-space: nowrap;
    border-bottom: 1px solid var(--border-subtle, #f0f0f2);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .data-row {
    transition: background-color 120ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .data-row td {
    padding: 14px 16px;
    vertical-align: middle;
    white-space: nowrap;
    border-bottom: 1px solid var(--border-subtle, #f0f0f2);
  }

  .data-row:hover {
    background: var(--surface-hover, #fafbfc);
  }

  .data-row.clickable {
    cursor: pointer;
  }

  .data-row.suspended {
    color: var(--ink-400, #9ca3af);
  }

  .data-row.suspended td {
    border-bottom: 1px dashed var(--border-subtle, #f0f0f2);
  }

  .cell {
    padding: 14px 16px;
  }

  .code {
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 13px;
    color: var(--ink-500, #6b7280);
    letter-spacing: 0.01em;
  }

  .name {
    font-weight: 500;
    color: var(--ink-900, #0f172a);
  }

  .badge {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    border-radius: 4px;
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 13px;
    font-weight: 600;
    line-height: 1.3;
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }

  .muted {
    color: var(--ink-400, #9ca3af);
  }

  .price {
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-variant-numeric: tabular-nums;
    text-align: right;
    color: var(--ink-700, #374151);
  }

  .pulse-dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--primary, #1e293b);
    animation: pulse 1.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.3; transform: scale(0.8); }
    50%      { opacity: 1;   transform: scale(1.2); }
  }

  .error-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--warning, #c8781a);
    color: #fff;
    font-size: 11px;
    font-weight: 700;
    cursor: help;
  }

  .suspended-badge {
    display: inline-flex;
    align-items: center;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    color: var(--ink-400, #9ca3af);
    background: var(--ink-50, #f9fafb);
    border: 1px solid var(--border-subtle, #f0f0f2);
  }

  .indicator {
    text-align: center;
    padding: 14px 8px;
  }
</style>
