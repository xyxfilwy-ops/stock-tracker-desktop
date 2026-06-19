<script lang="ts">
  import { onMount, tick } from 'svelte';

  export let visible: boolean = false;
  export let onConfirm: (code: string) => void;
  export let onCancel: () => void;
  export let error: string = '';
  export let loading: boolean = false;

  let inputValue: string = '';
  let inputRef: HTMLInputElement;

  $: if (visible) {
    tick().then(() => {
      if (inputRef) inputRef.focus();
    });
  }

  $: if (!visible) {
    inputValue = '';
  }

  function autoPrefix(code: string): string {
    const trimmed = code.trim().toLowerCase();
    if (trimmed.startsWith('sh') || trimmed.startsWith('sz')) {
      return trimmed;
    }
    if (/^6\d{5}$/.test(trimmed)) {
      return `sh${trimmed}`;
    }
    if (/^[03]\d{5}$/.test(trimmed)) {
      return `sz${trimmed}`;
    }
    return trimmed;
  }

  function handleConfirm() {
    if (!inputValue.trim() || loading) return;
    const code = autoPrefix(inputValue);
    onConfirm(code);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      handleConfirm();
    } else if (e.key === 'Escape') {
      onCancel();
    }
  }
</script>

{#if visible}
  <div class="dialog-overlay" on:click={onCancel}>
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <h2 class="dialog-title">选入股票</h2>
      </div>
      <div class="dialog-body">
        <div class="input-group">
          <label class="input-label" for="stock-code">股票代码</label>
          <input
            id="stock-code"
            bind:this={inputRef}
            bind:value={inputValue}
            on:keydown={handleKeydown}
            class="input"
            class:error={!!error}
            placeholder="如 600519 或 sh600519"
            type="text"
            autocomplete="off"
          />
          <p class="input-hint">支持 sh/sz 开头或纯数字代码，6 开头自动识别为上海，0/3 开头为深圳</p>
          {#if error}
            <p class="error-message">{error}</p>
          {/if}
        </div>
      </div>
      <div class="dialog-footer">
        <button class="btn btn-secondary" on:click={onCancel} type="button">
          取消
        </button>
        <button
          class="btn btn-primary"
          class:loading
          on:click={handleConfirm}
          disabled={loading || !inputValue.trim()}
          type="button"
        >
          {loading ? '获取中…' : '确认'}
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
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .input-label {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary, #1a1c23);
  }

  .input {
    height: 36px;
    padding: 0 12px;
    border: 1px solid var(--border, #e2e4e8);
    border-radius: 6px;
    font-family: var(--font-mono, 'JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', Consolas, monospace);
    font-size: 14px;
    color: var(--text-primary, #1a1c23);
    background: var(--surface, #ffffff);
    transition: border-color 200ms cubic-bezier(0.4, 0, 0.2, 1),
                box-shadow 200ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .input:focus {
    outline: none;
    border-color: var(--primary, #2563eb);
    box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.2);
  }

  .input.error {
    border-color: var(--warning, #d97706);
  }

  .input.error:focus {
    box-shadow: 0 0 0 2px rgba(217, 119, 6, 0.2);
  }

  .input-hint {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 12px;
    color: var(--text-muted, #9ca3af);
    margin: 0;
  }

  .error-message {
    font-family: var(--font-body, 'Inter', 'PingFang SC', 'Microsoft YaHei', system-ui, sans-serif);
    font-size: 13px;
    color: var(--warning, #d97706);
    margin: 0;
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

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: var(--primary, #2563eb);
    color: #ffffff;
  }

  .btn-secondary {
    background: var(--surface, #ffffff);
    color: var(--text-primary, #1a1c23);
    border: 1px solid var(--border, #e2e4e8);
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
