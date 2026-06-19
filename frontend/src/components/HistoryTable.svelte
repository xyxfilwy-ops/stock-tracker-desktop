<script lang="ts">
  import type { HistoryRecord } from '../services/api';

  export let history: HistoryRecord[] = [];

  function formatPrice(price: number): string {
    return (price / 100).toFixed(2);
  }

  function formatReturn(bp: number): string {
    const pct = (bp / 100).toFixed(2);
    return bp >= 0 ? `+${pct}%` : `${pct}%`;
  }

  function getReturnColor(bp: number): string {
    if (bp > 0) return 'var(--positive, #dc2626)';
    if (bp < 0) return 'var(--negative, #16a34a)';
    return 'var(--text-secondary, #6b7280)';
  }

  function formatDateRange(entry: string, exit: string): string {
    const entryFmt = formatShortDate(entry);
    const exitFmt = formatShortDate(exit);
    return `${entryFmt}→${exitFmt}`;
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
        <th class="cell" style="width: 80px;">代码</th>
        <th class="cell" style="width: 100px;">名称</th>
        <th class="cell" style="width: 100px; text-align: center;">选入→调出</th>
        <th class="cell" style="width: 60px; text-align: right;">天数</th>
        <th class="cell" style="width: 80px; text-align: right;">收益</th>
      </tr>
    </thead>
    <tbody>
      {#each history as record (record.id)}
        <tr class="data-row">
          <td class="cell code">{record.code}</td>
          <td class="cell name">{record.name}</td>
          <td class="cell" style="text-align: center;">
            {formatDateRange(record.entryDate, record.exitDate)}
          </td>
          <td class="cell" style="text-align: right;">
            {record.holdingDays}天
          </td>
          <td class="cell" style="text-align: right;">
            <span class="badge" style="color: {getReturnColor(record.totalReturn)};">
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
  }

  .history-table {
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
</style>