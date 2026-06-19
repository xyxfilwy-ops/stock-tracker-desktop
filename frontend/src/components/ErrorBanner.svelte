<script lang="ts">
  import { onMount, onDestroy } from 'svelte';

  export let type: 'toast' | 'banner' | 'offline' = 'toast';
  export let message: string = '';
  export let visible: boolean = false;
  export let onDismiss: (() => void) | undefined = undefined;
  export let onRetry: (() => void) | undefined = undefined;

  let toastTimer: ReturnType<typeof setTimeout> | null = null;

  $: if (type === 'toast' && visible && onDismiss) {
    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
      onDismiss?.();
    }, 5000);
  }

  $: if (!visible && toastTimer) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }

  onDestroy(() => {
    if (toastTimer) clearTimeout(toastTimer);
  });

  function handleDismiss() {
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    onDismiss?.();
  }
</script>

{#if visible && message}
  {#if type === 'toast'}
    <div class="error-toast" role="alert">
      <span class="toast-message">{message}</span>
      {#if onDismiss}
        <button class="toast-close" on:click={handleDismiss} type="button" aria-label="关闭">×</button>
      {/if}
    </div>
  {:else if type === 'banner'}
    <div class="error-banner" role="alert">
      <span class="banner-message">{message}</span>
    </div>
  {:else if type === 'offline'}
    <div class="error-offline" role="alert">
      <div class="offline-card">
        <div class="offline-icon">📡</div>
        <p class="offline-message">{message}</p>
        {#if onRetry}
          <button class="retry-btn" on:click={onRetry} type="button">重试</button>
        {/if}
      </div>
    </div>
  {/if}
{/if}

<style>
  .error-toast {
    position: fixed;
    top: 56px;
    left: 50%;
    transform: translateX(-50%);
    background: var(--surface, #ffffff);
    color: var(--text-primary, #1a1c23);
    padding: 10px 16px;
    border-radius: 6px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
    display: flex;
    align-items: center;
    gap: 12px;
    z-index: 999;
    animation: toastIn 200ms cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--border, #e2e4e8);
    max-width: 90vw;
  }

  .toast-message {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 13px;
  }

  .toast-close {
    background: none;
    border: none;
    color: var(--text-muted, #9ca3af);
    font-size: 18px;
    cursor: pointer;
    padding: 0 2px;
    line-height: 1;
    transition: color 150ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .toast-close:hover {
    color: var(--text-primary, #1a1c23);
  }

  .error-banner {
    background: var(--error-bg, #fef2f2);
    color: var(--text-primary, #1a1c23);
    padding: 10px 16px;
    text-align: center;
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 13px;
    border-bottom: 1px solid var(--border, #e2e4e8);
  }

  .banner-message {
    margin: 0;
  }

  .error-offline {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 48px 16px;
    flex: 1;
  }

  .offline-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 12px;
    background: var(--surface, #ffffff);
    border: 1px solid var(--border, #e2e4e8);
    border-radius: 12px;
    padding: 32px 24px;
    max-width: 320px;
  }

  .offline-icon {
    font-size: 48px;
    line-height: 1;
    color: var(--text-muted, #9ca3af);
  }

  .offline-message {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 14px;
    color: var(--text-primary, #1a1c23);
    margin: 0;
  }

  .retry-btn {
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

  .retry-btn:hover {
    opacity: 0.9;
  }

  @keyframes toastIn {
    from { opacity: 0; transform: translateX(-50%) translateY(-8px); }
    to { opacity: 1; transform: translateX(-50%) translateY(0); }
  }
</style>
