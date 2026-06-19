<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { SearchStocks, type SearchResult } from '../services/api';

  export let visible: boolean = false;
  export let onConfirm: (code: string) => void;
  export let onCancel: () => void;
  export let error: string = '';
  export let loading: boolean = false;

  let inputValue: string = '';
  let inputRef: HTMLInputElement;
  let searchResults: SearchResult[] = [];
  let searchLoading: boolean = false;
  let selectedIndex: number = -1;
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  $: if (visible) {
    tick().then(() => {
      if (inputRef) inputRef.focus();
    });
  }

  $: if (!visible) {
    inputValue = '';
    searchResults = [];
    selectedIndex = -1;
    searchLoading = false;
    if (searchTimer) {
      clearTimeout(searchTimer);
      searchTimer = null;
    }
  }

  function isPureDigits(str: string): boolean {
    return /^\d+$/.test(str.trim());
  }

  function isPrefixedCode(str: string): boolean {
    const trimmed = str.trim().toLowerCase();
    return /^s[hz]\d{6}$/.test(trimmed);
  }

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    const val = target.value.trim();
    inputValue = target.value;
    searchResults = [];
    selectedIndex = -1;

    if (!val) return;

    // 只有带前缀的代码（如 sh600519）不搜索，直接确认
    if (isPrefixedCode(val)) {
      return;
    }

    // 其他情况都触发搜索（包括纯数字如 000001、名称、首字母）
    if (searchTimer) clearTimeout(searchTimer);
    searchLoading = true;
    searchTimer = setTimeout(async () => {
      try {
        const results = await SearchStocks(val);
        searchResults = results || [];
        selectedIndex = searchResults.length > 0 ? 0 : -1;
      } catch (err) {
        searchResults = [];
        selectedIndex = -1;
      } finally {
        searchLoading = false;
      }
    }, 250);
  }

  function selectResult(result: SearchResult) {
    inputValue = result.name;
    searchResults = [];
    selectedIndex = -1;
    onConfirm(result.code);
  }

  function handleConfirm() {
    if (!inputValue.trim() || loading) return;
    const val = inputValue.trim();

    // 如果有候选结果，选中第一个
    if (searchResults.length > 0 && selectedIndex >= 0) {
      selectResult(searchResults[selectedIndex]);
      return;
    }

    // 如果输入带前缀的代码，直接传入
    if (isPrefixedCode(val)) {
      onConfirm(val);
      return;
    }

    // 其他情况直接传原始值（让后端 normalize）
    onConfirm(val);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (searchResults.length > 0) {
        selectedIndex = (selectedIndex + 1) % searchResults.length;
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (searchResults.length > 0) {
        selectedIndex = selectedIndex <= 0 ? searchResults.length - 1 : selectedIndex - 1;
      }
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (searchResults.length > 0 && selectedIndex >= 0) {
        selectResult(searchResults[selectedIndex]);
      } else {
        handleConfirm();
      }
    } else if (e.key === 'Escape') {
      if (searchResults.length > 0) {
        searchResults = [];
        selectedIndex = -1;
      } else {
        onCancel();
      }
    }
  }

  function typeLabel(type: string): string {
    switch (type) {
      case 'fund': return '场内基金';
      case 'otc_fund': return '场外基金';
      default: return '股票';
    }
  }
</script>

{#if visible}
  <div class="dialog-overlay">
    <div class="dialog" on:keydown={handleKeydown} tabindex="-1" role="dialog" aria-modal="true">
      <div class="dialog-header">
        <h2 class="dialog-title">选入股票/基金</h2>
        <button class="dialog-close" on:click={onCancel} type="button" aria-label="关闭">×</button>
      </div>
      <div class="dialog-body">
        <div class="dialog-input-group">
          <label class="dialog-input-label" for="stock-code">股票/基金代码或名称</label>
          <div class="input-wrapper">
            <input
              id="stock-code"
              bind:this={inputRef}
              bind:value={inputValue}
              on:input={handleInput}
              on:keydown={handleKeydown}
              class="input"
              class:error={!!error}
              placeholder="如 600519、贵州茅台 或 gzmt"
              type="text"
              autocomplete="off"
            />
            {#if searchLoading}
              <span class="search-spinner">⟳</span>
            {/if}
          </div>
          <p class="dialog-input-hint">
            支持：股票代码（6位数字）、名称、拼音首字母（如 gzmt）
          </p>
          {#if error}
            <p class="dialog-error">{error}</p>
          {/if}

          <!-- 搜索候选列表 -->
          {#if searchResults.length > 0}
            <ul class="search-results" role="listbox" aria-label="搜索结果">
              {#each searchResults as result, i}
                <li
                  role="option"
                  aria-selected={i === selectedIndex}
                  class="search-result-item"
                  class:selected={i === selectedIndex}
                  on:click={() => selectResult(result)}
                  on:mouseenter={() => selectedIndex = i}
                >
                  <span class="result-name">{result.name}</span>
                  <span class="result-meta">
                    <span class="result-type">{typeLabel(result.type)}</span>
                    <span class="result-code">{result.code}</span>
                  </span>
                </li>
              {/each}
            </ul>
          {:else if inputValue.trim() && !searchLoading && !isPrefixedCode(inputValue) && !/^\d{6}$/.test(inputValue.trim())}
            <p class="search-no-results">未找到匹配结果，请直接输入代码</p>
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
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
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
    width: 480px;
    max-width: calc(100vw - 32px);
    background: var(--surface, #ffffff);
    border-radius: var(--radius-lg, 12px);
    padding: var(--space-32, 32px);
    box-shadow: var(--shadow-xl, 0 16px 48px rgba(0,0,0,0.08));
    animation: dialogEnter 300ms cubic-bezier(0.4, 0, 0.2, 1);
    overflow: visible;
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
    position: relative;
  }

  .dialog-input-label {
    font-size: 13px;
    font-weight: 500;
    color: var(--ink-700, #374151);
  }

  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }

  .input {
    height: 42px;
    width: 100%;
    padding: 0 var(--space-16, 16px);
    padding-right: 40px;
    border: 1px solid var(--border, #e5e7eb);
    border-radius: var(--radius-md, 8px);
    font-family: var(--font-body, 'Inter', sans-serif);
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

  .search-spinner {
    position: absolute;
    right: 12px;
    color: var(--ink-400, #9ca3af);
    font-size: 16px;
    animation: spin 1s linear infinite;
    pointer-events: none;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
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

  .search-results {
    list-style: none;
    margin: 4px 0 0 0;
    padding: 4px;
    max-height: 200px;
    overflow-y: auto;
    border: 1px solid var(--border, #e5e7eb);
    border-radius: var(--radius-md, 8px);
    background: var(--surface, #ffffff);
    box-shadow: var(--shadow-lg, 0 8px 24px rgba(0,0,0,0.06));
  }

  .search-result-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    transition: background 120ms ease;
    gap: 8px;
  }

  .search-result-item:hover,
  .search-result-item.selected {
    background: var(--ink-50, #f9fafb);
  }

  .result-name {
    font-size: 14px;
    font-weight: 500;
    color: var(--ink-900, #0f172a);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .result-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .result-type {
    font-size: 11px;
    font-weight: 500;
    padding: 2px 6px;
    border-radius: 4px;
    background: var(--ink-100, #f3f4f6);
    color: var(--ink-500, #6b7280);
  }

  .result-code {
    font-size: 12px;
    font-family: var(--font-mono, 'SF Mono', monospace);
    color: var(--ink-400, #9ca3af);
  }

  .search-no-results {
    font-size: 13px;
    color: var(--ink-400, #9ca3af);
    margin: 8px 0 0 0;
    padding: 8px 12px;
    background: var(--ink-50, #f9fafb);
    border-radius: var(--radius-sm, 6px);
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
