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
    if (bp > 0) return 'var(--positive, #b9403a)';
    if (bp < 0) return 'var(--negative, #2e8b57)';
    return 'var(--neutral, #6b7280)';
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
  <div class="dialog-overlay">
    <div class="dialog" role="dialog" aria-modal="true">
      <div class="dialog-header">
        <h2 class="dialog-title">确认调出</h2>
        <button class="dialog-close" on:click={onCancel} type="button" aria-label="关闭">×</button>
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
        <div class="info-row" style="border-bottom: none;">
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
    background: rgba(15, 23, 42, 0.35);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 200ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .dialog {
    width: 420px;
    max-width: calc(100vw - 32px);
    background: var(--surface, #ffffff);
    border-radius: var(--radius-lg, 12px);
    padding: var(--space-32, 32px);
    box-shadow: var(--shadow-xl, 0 16px 48px rgba(0,0,0,0.08));
    animation: dialogEnter 300ms cubic-bezier(0.4, 0, 0.2, 1);
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-24, 24px);
    border: 1px solid var(--border-subtle, #f0f0f2);
  }

  .dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .dialog-title {
    font-size: 18px;
    font-weight: 700;
    color: var(--ink-900, #0f172a);
    letter-spacing: -0.02em;
    margin: 0;
  }

  .dialog-close {
    width: 32px;
    height: 32px;
    border-radius: 999px;
    border: none;
    background: transparent;
    color: var(--ink-400, #9ca3af);
    font-size: 20px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 120ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .dialog-close:hover {
    background: var(--ink-50, #f9fafb);
    color: var(--ink-700, #374151);
  }

  .dialog-body {
    font-size: 14px;
    color: var(--ink-700, #374151);
    line-height: 1.6;
    display: flex;
    flex-direction: column;
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-12, 12px);
    padding-top: var(--space-8, 8px);
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 0;
    border-bottom: 1px solid var(--border-subtle, #f0f0f2);
  }

  .info-row:last-child {
    border-bottom: none;
  }

  .info-label {
    font-size: 13px;
    color: var(--ink-500, #6b7280);
    font-weight: 400;
  }

  .info-value {
    font-size: 14px;
    font-weight: 500;
    color: var(--ink-900, #0f172a);
  }

  .info-value.code {
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 13px;
  }

  .info-value.return {
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-weight: 700;
    font-size: 15px;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 38px;
    padding: 0 var(--space-16, 16px);
    border-radius: var(--radius-md, 8px);
    font-family: var(--font-body, 'Inter', sans-serif);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    border: none;
    outline: none;
    transition: all 120ms cubic-bezier(0.4, 0, 0.2, 1);
    white-space: nowrap;
    gap: var(--space-8, 8px);
  }

  .btn-secondary {
    background: var(--ink-50, #f9fafb);
    color: var(--ink-700, #374151);
    border: 1px solid var(--border, #e5e7eb);
  }
  .btn-secondary:hover {
    background: var(--ink-100, #f3f4f6);
  }

  .btn-danger {
    background: var(--positive, #b9403a);
    color: #ffffff;
  }
  .btn-danger:hover {
    background: #a33530;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes dialogEnter {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
</style>
