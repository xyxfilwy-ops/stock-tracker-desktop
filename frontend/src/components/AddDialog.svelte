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
  <div class="dialog-overlay">
    <div class="dialog" on:keydown={handleKeydown} tabindex="-1" role="dialog" aria-modal="true">
      <div class="dialog-header">
        <h2 class="dialog-title">选入股票</h2>
        <button class="dialog-close" on:click={onCancel} type="button" aria-label="关闭">×</button>
      </div>
      <div class="dialog-body">
        <div class="dialog-input-group">
          <label class="dialog-input-label" for="stock-code">股票代码</label>
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
          <p class="dialog-input-hint">支持 sh/sz 开头或纯数字代码，6 开头自动识别为上海，0/3 开头为深圳</p>
          {#if error}
            <p class="dialog-error">{error}</p>
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
          {loading ? '获取中…' : '确认选入'}
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
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-12, 12px);
    padding-top: var(--space-8, 8px);
  }

  .dialog-input-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-8, 8px);
  }

  .dialog-input-label {
    font-size: 13px;
    font-weight: 500;
    color: var(--ink-700, #374151);
  }

  .dialog-input-hint {
    font-size: 12px;
    color: var(--ink-400, #9ca3af);
    line-height: 1.5;
    margin: 0;
  }

  .dialog-error {
    font-size: 13px;
    color: var(--positive, #b9403a);
    margin: var(--space-4, 4px) 0 0 0;
  }

  .input {
    height: 42px;
    width: 100%;
    padding: 0 var(--space-16, 16px);
    border: 1px solid var(--border, #e5e7eb);
    border-radius: var(--radius-md, 8px);
    font-family: var(--font-mono, 'SF Mono', monospace);
    font-size: 14px;
    color: var(--ink-700, #374151);
    background: var(--surface, #ffffff);
    outline: none;
    transition: border-color 200ms cubic-bezier(0.4, 0, 0.2, 1),
                box-shadow 200ms cubic-bezier(0.4, 0, 0.2, 1);
  }

  .input::placeholder {
    color: var(--ink-400, #9ca3af);
    font-family: var(--font-body, 'Inter', sans-serif);
  }

  .input:focus {
    border-color: var(--ink-900, #0f172a);
    box-shadow: 0 0 0 3px rgba(30, 41, 59, 0.08);
  }

  .input.error {
    border-color: var(--positive, #b9403a);
  }

  .input.error:focus {
    box-shadow: 0 0 0 3px rgba(185, 64, 58, 0.1);
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

  .btn-primary {
    background: var(--primary, #1e293b);
    color: #fff;
  }
  .btn-primary:hover:not(:disabled) {
    background: var(--primary-hover, #0f172a);
  }

  .btn-secondary {
    background: var(--ink-50, #f9fafb);
    color: var(--ink-700, #374151);
    border: 1px solid var(--border, #e5e7eb);
  }
  .btn-secondary:hover:not(:disabled) {
    background: var(--ink-100, #f3f4f6);
  }

  .btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    transform: none !important;
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
