<script lang="ts">
  export let loading: boolean = false;
  export let progress: { current: number; total: number } | null = null;

  $: progressPercent = progress && progress.total > 0
    ? Math.round((progress.current / progress.total) * 100)
    : 0;
</script>

{#if loading}
  <div class="progress-bar" style="width: {progressPercent}%; transition: width 300ms cubic-bezier(0.4, 0, 0.2, 1);"></div>
{/if}

<style>
  .progress-bar {
    position: absolute;
    bottom: 0;
    left: 0;
    height: 2px;
    background: var(--primary, #2563eb);
    z-index: 10;
  }
</style>
