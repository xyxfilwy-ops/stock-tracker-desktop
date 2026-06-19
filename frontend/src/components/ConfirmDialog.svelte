<script lang="ts">
  import type { Stock } from '../services/api';

  export let visible: boolean = false;
  export let stock: Stock | null = null;
  export let onConfirm: () => void;
  export let onCancel: () => void;

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

  function estimatedReturn(): number {
    if (!stock) return 0;
    return stock.currentPrice - stock.entryPrice;
  }

  function estimatedReturnPct(): number {
    if (!stock || stock.entryPrice === 0) return 0;
    return Math.round(((stock.currentPrice - stock.entryPrice) / stock.entryPrice) * 10000);
  }
</script>

{#if visible && stock}
  <div class="dialog-overlay" on:click={onCancel}>
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <h2 class="dialog-title">确认调出</h2>
      </div>
      <div class="dialog-body">
        <div class="info-row">
          <span class="info-label">股票代码</span>
          <span class="info-value code">{stock.code}</span>
        </div>
        <div class="info-row">
          <span class="info-label">股票名称</span>
          <span class="info-value">{stock.name}</span>
        </div>
        <div class="info-row">
          <span class="info-label">选入价</span>
          <span class="info-value">{formatPrice(stock.entryPrice)} 元</span>
        </div>
        <div class="info-row">
          <span class="info-label">当前价</span>
          <span class="info-value">{formatPrice(stock.currentPrice)} 元</span>
        </div>
        <div class="info-row">
          <span class="info-label">预估收益</span>
          <span class="info-value return" style="color: {getChangeColor(estimatedReturnPct())};">
            {estimatedReturn() >= 0 ? '+' : ''}{formatPrice(Math.abs(estimatedReturn()))} 元
            ({formatChange(estimatedReturnPct())})
          </span>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="btn btn-secondary" on:click={onCancel} type="button">
          取消
        </button>
        <button class="btn btn-danger" on:click={onConfirm} type="button">
          确认调出
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 200ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .dialog {
    width: 400px;
    background: var(--surface, #ffffff);
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
    animation: dialogEnter 250ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .dialog-header {
    margin-bottom: 20px;
  }

  .dialog-title {
    font-family: var(--font-display, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary, #1a1c23);
    margin: 0;
  }

  .dialog-body {
    margin-bottom: 24px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .info-label {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 14px;
    color: var(--text-secondary, #6b7280);
  }

  .info-value {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary, #1a1c23);
  }

  .code {
    font-family: var(--font-mono, 'JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', Consolas, monospace);
  }

  .return {
    font-family: var(--font-mono, 'JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', Consolas, monospace);
    font-weight: 600;
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .btn {
    height: 36px;
    padding: 0 16px;
    border-radius: 6px;
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    border: none;
    transition: opacity 150ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .btn:hover:not(:disabled) {
    opacity: 0.9;
  }

  .btn-secondary {
    background: var(--surface, #ffffff);
    color: var(--text-primary, #1a1c23);
    border: 1px solid var(--border, #e2e4e8);
  }

  .btn-danger {
    background: var(--negative, #16a34a);
    color: #ffffff;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes dialogEnter {
    from { opacity: 0; transform: translateY(-8px); }
    to { opacity: 1; transform: translateY(0); }
  }
</style>
