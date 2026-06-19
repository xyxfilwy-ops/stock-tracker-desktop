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

  function formatDateRange(entry: string, exit: string): string {
    const entryFmt = formatShortDate(entry);
    const exitFmt = formatShortDate(exit);
    return `${entryFmt} → ${exitFmt}`;
  }

  function formatShortDate(dateStr: string): string {
    const d = new Date(dateStr);
    return `${d.getMonth() + 1}/${String(d.getDate()).padStart(2, '0')}`;
  }
</script>

<div class="table-wrapper">
  <table class="history-table">
    <thead>
      <tr class="header-row">
        <th class="cell">代码</th>
        <th class="cell">名称</th>
        <th class="cell" style="text-align: center;">选入 → 调出</th>
        <th class="cell" style="text-align: right;">天数</th>
        <th class="cell" style="text-align: right;">收益</th>
      </tr>
    </thead>
    <tbody>
      {#each history as record (record.id)}
        <tr class="data-row">
          <td class="cell code">{record.code}</td>
          <td class="cell name">{record.name}</td>
          <td class="cell" style="text-align: center;">
            <span class="date-range">{formatDateRange(record.entryDate, record.exitDate)}</span>
          </td>
          <td class="cell" style="text-align: right;">
            <span class="days">{record.holdingDays}天</span>
          </td>
          <td class="cell" style="text-align: right;">
            <span class="badge" style="color: {getReturnColor(record.totalReturn)}; background: {record.totalReturn > 0 ? 'var(--positive-bg)' : record.totalReturn < 0 ? 'var(--negative-bg)' : 'rgba(107,114,128,0.06)'};">
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

  .date-range {
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 13px;
    color: var(--ink-500, #6b7280);
    letter-spacing: 0.01em;
  }

  .days {
    font-size: 13px;
    color: var(--ink-500, #6b7280);
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
