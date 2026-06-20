<script lang="ts">
  import type { HistoryRecord } from '../services/api';

  export let history: HistoryRecord[] = [];
  export let onClear: () => void = () => {};

  function formatPrice(price: number): string {
    return (price / 100).toFixed(2);
  }

  function formatReturn(bp: number): string {
    const pct = (bp / 100).toFixed(2);
    return bp >= 0 ? `+${pct}%` : `${pct}%`;
  }

  function getReturnColor(bp: number): string {
    if (bp > 0) return 'var(--positive, #b9403a)';
    if (bp < 0) return 'var(--negative, #2e8b57)';
    return 'var(--neutral, #6b7280)';
  }

  function getReturnBg(bp: number): string {
    if (bp > 0) return 'rgba(185, 64, 58, 0.08)';
    if (bp < 0) return 'rgba(46, 139, 87, 0.08)';
    return 'rgba(107, 114, 128, 0.06)';
  }

  function formatShortDate(dateStr: string): string {
    const d = new Date(dateStr);
    return `${d.getMonth() + 1}/${String(d.getDate()).padStart(2, '0')}`;
  }

  function formatDateTime(date: string, time: string): string {
    return `${date} ${time}`;
  }
</script>

<div class="table-wrapper">
  <table class="history-table">
    <thead>
      <tr class="header-row">
        <th class="cell">代码</th>
        <th class="cell">名称</th>
        <th class="cell" style="text-align: right;">选入价</th>
        <th class="cell" style="text-align: right;">调出价</th>
        <th class="cell" style="text-align: center;">持仓时间</th>
        <th class="cell" style="text-align: right;">收益</th>
      </tr>
    </thead>
    <tbody>
      {#each history as record (record.id)}
        <tr class="data-row">
          <td class="cell code">{record.code}</td>
          <td class="cell name">{record.name}</td>
          <td class="cell" style="text-align: right;">
            <span class="price">{formatPrice(record.entryPrice)}</span>
            <span class="time-hint">{formatDateTime(record.entryDate, record.entryTime)}</span>
          </td>
          <td class="cell" style="text-align: right;">
            <span class="price">{formatPrice(record.exitPrice)}</span>
            <span class="time-hint">{formatDateTime(record.exitDate, record.exitTime)}</span>
          </td>
          <td class="cell" style="text-align: center;">
            <span class="duration">{record.holdingDuration}</span>
            <span class="days-hint">({record.holdingDays}天)</span>
          </td>
          <td class="cell" style="text-align: right;">
            <span class="badge" style="color: {getReturnColor(record.totalReturn)}; background: {getReturnBg(record.totalReturn)};">
              {formatReturn(record.totalReturn)}
            </span>
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

  .history-table {
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
    padding: 10px 10px;
    white-space: nowrap;
    border-bottom: 1px solid var(--border-subtle, #f0f0f2);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  .data-row {
    transition: background-color 120ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .data-row td {
    padding: 10px 10px;
    vertical-align: middle;
    white-space: nowrap;
    border-bottom: 1px solid var(--border-subtle, #f0f0f2);
  }

  .data-row:hover {
    background: var(--surface-hover, #fafbfc);
  }

  .cell {
    padding: 10px 10px;
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

  .price {
    display: block;
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 13px;
    color: var(--ink-700, #374151);
    font-variant-numeric: tabular-nums;
  }

  .time-hint {
    display: block;
    font-size: 11px;
    color: var(--ink-400, #9ca3af);
    margin-top: 2px;
  }

  .duration {
    font-size: 13px;
    color: var(--ink-700, #374151);
  }

  .days-hint {
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
</style>
