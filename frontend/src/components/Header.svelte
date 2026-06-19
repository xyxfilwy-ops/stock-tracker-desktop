<script lang="ts">
  import RefreshIndicator from './RefreshIndicator.svelte';

  export let loading: boolean = false;
  export let onRefresh: () => void;
</script>

<header class="header">
  <div class="header-inner">
    <div class="title">StockTracker</div>
    <button
      class="refresh-btn"
      class:loading
      on:click={onRefresh}
      disabled={loading}
      type="button"
    >
      <span class="refresh-icon" class:spinning={loading}>⟳</span>
      <span class="refresh-text">
        {loading ? '刷新中…' : '刷新'}
      </span>
    </button>
  </div>
  <RefreshIndicator {loading} />
</header>

<style>
  .header {
    height: 48px;
    background: var(--surface, #ffffff);
    border-bottom: 1px solid var(--border, #e2e4e8);
    position: relative;
    flex-shrink: 0;
  }

  .header-inner {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;
    max-width: 720px;
    margin: 0 auto;
  }

  .title {
    font-family: var(--font-display, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 20px;
    font-weight: 600;
    color: var(--text-primary, #1a1c23);
    line-height: 1;
  }

  .refresh-btn {
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
    display: flex;
    align-items: center;
    gap: 6px;
    transition: opacity 150ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .refresh-btn:hover:not(:disabled) {
    opacity: 0.9;
  }

  .refresh-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .refresh-icon {
    display: inline-block;
    font-size: 14px;
    line-height: 1;
  }

  .refresh-icon.spinning {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
