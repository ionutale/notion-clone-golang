<script lang="ts">
  import { getDeletedBlock, clearToast } from '$lib/stores/toast.svelte';
  import { blockStore } from '$lib/stores/blocks.svelte';

  let deleted = $derived(getDeletedBlock());

  async function handleUndo() {
    if (!deleted) return;
    await blockStore.restoreBlock(deleted.id);
    clearToast();
  }

  function handleDismiss() {
    clearToast();
  }
</script>

{#if deleted}
  <div class="toast toast-bottom toast-center z-50">
    <div class="alert alert-info shadow-lg flex items-center gap-3 px-4 py-3">
      <span class="text-sm">Deleted</span>
      <button onclick={handleUndo} class="btn btn-ghost btn-xs font-semibold">Undo</button>
      <button onclick={handleDismiss} class="btn btn-ghost btn-xs" aria-label="Dismiss">
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  </div>
{/if}
