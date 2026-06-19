<script lang="ts">
  export let title: string = '还没有跟踪任何股票';
  export let description: string = '点击「+ 选入股票」开始跟踪';
  export let hint: string = '支持 sh/sz 开头或纯数字代码';
  export let onAdd: (() => void) | undefined = undefined;
</script>

<div class="empty-state">
  <div class="empty-state-icon">
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
    </svg>
  </div>
  <h3 class="empty-state-title">{title}</h3>
  <p class="empty-state-desc">{description}</p>
  {#if hint}
    <p class="empty-state-hint">{hint}</p>
  {/if}
  {#if onAdd}
    <button class="add-btn" on:click={onAdd} type="button">
      + 选入股票
    </button>
  {/if}
</div>

<style>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: var(--space-48, 48px) var(--space-16, 16px);
    text-align: center;
    gap: var(--space-12, 12px);
  }

  .empty-state-icon {
    width: 56px;
    height: 56px;
    border-radius: var(--radius-lg, 12px);
    background: var(--ink-50, #f9fafb);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--ink-400, #9ca3af);
    margin-bottom: var(--space-8, 8px);
  }

  .empty-state-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--ink-900, #0f172a);
    letter-spacing: -0.01em;
    margin: 0;
  }

  .empty-state-desc {
    font-size: 14px;
    color: var(--ink-500, #6b7280);
    line-height: 1.6;
    max-width: 320px;
    margin: 0;
  }

  .empty-state-hint {
    font-size: 12px;
    color: var(--ink-400, #9ca3af);
    font-family: var(--font-mono, 'SF Mono', monospace);
    margin: 0;
  }

  .add-btn {
    margin-top: var(--space-8, 8px);
    height: 38px;
    padding: 0 var(--space-20, 20px);
    border-radius: var(--radius-md, 8px);
    border: none;
    background: var(--primary, #1e293b);
    color: #ffffff;
    font-family: var(--font-body, 'Inter', sans-serif);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 120ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .add-btn:hover {
    background: var(--primary-hover, #0f172a);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md, 0 4px 12px rgba(0,0,0,0.04));
  }
</style>
