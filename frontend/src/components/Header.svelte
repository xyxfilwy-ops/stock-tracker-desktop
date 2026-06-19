<script lang="ts">
  import RefreshIndicator from './RefreshIndicator.svelte';
</script>

<header class="header">
  <div class="header-inner">
    <div>
      <div class="title">StockTracker</div>
      <div class="subtitle">本地股票跟踪 · 记录完整生命周期</div>
    </div>
  </div>
  <RefreshIndicator loading={false} />
</header>

<style>
  .header {
    height: 56px;
    position: relative;
    flex-shrink: 0;
  }

  .header-inner {
    height: 100%;
    display: flex;
    align-items: center;
    padding: 0 4px;
  }

  .title {
    font-family: var(--font-display, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 22px;
    font-weight: 700;
    color: var(--text-primary, #1a1c23);
    line-height: 1.2;
    letter-spacing: -0.03em;
  }

  .subtitle {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 12px;
    color: var(--text-muted, #9ca3af);
    font-weight: 400;
    margin-top: 2px;
  }
</style>
