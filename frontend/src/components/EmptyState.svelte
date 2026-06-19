<script lang="ts">
  export let title: string = '还没有跟踪任何股票';
  export let description: string = '点击「+ 选入股票」开始跟踪';
  export let hint: string = '支持 sh/sz 开头或纯数字代码';
  export let onAdd: (() => void) | undefined = undefined;
</script>

<div class="empty-state">
  <div class="empty-card">
    <div class="empty-icon">📋</div>
    <h3 class="empty-title">{title}</h3>
    <p class="empty-desc">{description}</p>
    {#if hint}
      <p class="empty-hint">{hint}</p>
    {/if}
    <p class="empty-example">例如：600519 或 sh600519</p>
    {#if onAdd}
      <button class="add-btn" on:click={onAdd} type="button">
        + 选入股票
      </button>
    {/if}
  </div>
</div>

<style>
  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 48px 16px;
    flex: 1;
  }

  .empty-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 8px;
  }

  .empty-icon {
    font-size: 48px;
    line-height: 1;
    color: var(--text-muted, #9ca3af);
    margin-bottom: 8px;
  }

  .empty-title {
    font-family: var(--font-display, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary, #1a1c23);
    margin: 0;
  }

  .empty-desc {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 14px;
    color: var(--text-secondary, #6b7280);
    margin: 0;
  }

  .empty-hint {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 13px;
    color: var(--text-muted, #9ca3af);
    margin: 4px 0 0 0;
  }

  .empty-example {
    font-family: var(--font-mono, 'JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', Consolas, monospace);
    font-size: 12px;
    color: var(--text-muted, #9ca3af);
    margin: 0;
  }

  .add-btn {
    margin-top: 16px;
    height: 36px;
    padding: 0 16px;
    border-radius: 6px;
    border: none;
    background: var(--primary, #2563eb);
    color: #ffffff;
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 150ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .add-btn:hover {
    opacity: 0.9;
  }
</style>
