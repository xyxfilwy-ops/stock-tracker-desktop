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

  function formatEntryDate(dateStr: string): string {
    const d = new Date(dateStr);
    const month = d.getMonth() + 1;
    const day = String(d.getDate()).padStart(2, '0');
    return `${month}/${day}`;
  }

  function isErrorRow(stock: Stock): boolean {
    return stock.status === 'normal' && stock.lastUpdate === undefined && stock.dataSource === undefined;
  }

  function handleRemoveClick(stock: Stock) {
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
        <th class="cell" style="text-align: right;">选入日期</th>
        <th class="cell" style="text-align: center;">时间</th>
        <th class="cell" style="text-align: right;">选入价</th>
        <th class="cell" style="text-align: right;">现价</th>
        <th class="cell" style="text-align: right;">日涨跌</th>
        <th class="cell" style="text-align: right;">累计涨跌</th>
        <th class="cell" style="text-align: center; width: 60px;">操作</th>
      </tr>
    </thead>
    <tbody>
      {#each stocks as stock (stock.id)}
        <tr
          class="data-row"
          class:suspended={stock.status === 'suspended'}
          class:error={isErrorRow(stock)}
        >
          <td class="cell code">{stock.code}</td>
          <td class="cell name">{stock.name}</td>
          <td class="cell" style="text-align: right;">
            <span class="date-text">{formatEntryDate(stock.entryDate)}</span>
          </td>
          <td class="cell" style="text-align: center;">
            <span class="time-text">{stock.entryTime}</span>
          </td>
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
              <span class="badge" style="color: {getChangeColor(stock.dailyChange)}; background: {stock.dailyChange > 0 ? 'var(--positive-bg)' : stock.dailyChange < 0 ? 'var(--negative-bg)' : 'rgba(107,114,128,0.06)'};">
                {formatChange(stock.dailyChange)}
              </span>
            {/if}
          </td>
          <td class="cell" style="text-align: right;">
            {#if stock.status === 'suspended'}
              <span class="muted">—</span>
            {:else}
              <span class="badge" style="color: {getChangeColor(stock.accChange)}; background: {stock.accChange > 0 ? 'var(--positive-bg)' : stock.accChange < 0 ? 'var(--negative-bg)' : 'rgba(107,114,128,0.06)'};">
                {formatChange(stock.accChange)}
              </span>
            {/if}
          </td>
          <td class="cell" style="text-align: center;">
            {#if stock.status === 'suspended'}
              <span class="suspended-badge">停牌</span>
            {:else}
              <button
                class="remove-btn"
                on:click={() => handleRemoveClick(stock)}
                title="调出"
                type="button"
              >
                调出
              </button>
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

  .date-text {
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 12px;
    color: var(--ink-500, #6b7280);
  }

  .time-text {
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 12px;
    color: var(--ink-400, #9ca3af);
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

  .remove-btn {
    height: 28px;
    padding: 0 12px;
    border-radius: 6px;
    border: 1px solid var(--border, #e5e7eb);
    background: var(--surface, #ffffff);
    color: var(--ink-500, #6b7280);
    font-family: var(--font-body, 'Inter', sans-serif);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 120ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .remove-btn:hover {
    background: var(--positive, #b9403a);
    color: #ffffff;
    border-color: var(--positive, #b9403a);
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
</style>
