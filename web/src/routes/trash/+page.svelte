<script lang="ts">
  import { api } from '$lib/api';
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { goto } from '$app/navigation';
  import type { PageSummary } from '$lib/types';
  import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';

  let pages = $state.raw<PageSummary[]>([]);
  let loading = $state(true);
  let deleteConfirmId = $state<string | null>(null);
  let showEmptyTrashConfirm = $state(false);

  async function loadTrash() {
    loading = true;
    try {
      pages = await api.listTrash();
    } catch (e) {
      console.error('Failed to load trash', e);
    } finally {
      loading = false;
    }
  }

  async function restore(id: string) {
    const restored = await blockStore.restoreBlock(id);
    pages = pages.filter(p => p.id !== id);
  }

  async function permanentDelete(id: string) {
    try {
      await api.permanentDeleteBlock(id);
      pages = pages.filter(p => p.id !== id);
    } catch (e) {
      console.error('Failed to permanently delete', e);
    }
    deleteConfirmId = null;
  }

  async function emptyTrash() {
    for (const p of [...pages]) {
      await permanentDelete(p.id);
    }
    showEmptyTrashConfirm = false;
  }

  $effect(() => { loadTrash(); });
</script>

<ConfirmDialog
  open={deleteConfirmId !== null}
  title="Delete forever?"
  message="Delete this page forever? This cannot be undone."
  confirmText="Delete forever"
  variant="danger"
  onConfirm={() => { if (deleteConfirmId) permanentDelete(deleteConfirmId); }}
  onCancel={() => deleteConfirmId = null}
/>

<ConfirmDialog
  open={showEmptyTrashConfirm}
  title="Empty trash?"
  message={`Delete all ${pages.length} pages forever? This cannot be undone.`}
  confirmText="Empty trash"
  variant="danger"
  onConfirm={emptyTrash}
  onCancel={() => showEmptyTrashConfirm = false}
/>

<div class="max-w-3xl mx-auto py-8 px-4">
  <div class="flex items-center justify-between mb-6">
    <h1 class="text-2xl font-bold">Trash</h1>
    {#if pages.length > 0}
      <button onclick={() => showEmptyTrashConfirm = true} class="btn btn-ghost btn-sm text-error">Empty trash</button>
    {/if}
  </div>

  {#if loading}
    <div class="flex justify-center py-20">
      <span class="loading loading-spinner loading-lg text-primary"></span>
    </div>
  {:else if pages.length === 0}
    <div class="text-center py-20 text-base-content/40">
      <p class="text-lg">Trash is empty</p>
      <p class="text-sm mt-1">Deleted pages will appear here</p>
    </div>
  {:else}
    <div class="space-y-1">
      {#each pages as p (p.id)}
        <div class="flex items-center gap-3 px-3 py-2 rounded-lg hover:bg-base-200 transition-colors group">
          {#if p.icon_type === 'image'}
            <img src={p.icon} alt="" class="w-4 h-4 rounded object-cover shrink-0" />
          {:else if p.icon}
            <span class="text-sm shrink-0">{p.icon}</span>
          {:else}
            <svg class="w-4 h-4 shrink-0 text-base-content/40" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          {/if}
          <div class="flex-1 min-w-0">
            <p class="text-sm truncate">{p.title}</p>
            <p class="text-xs text-base-content/40">Deleted</p>
          </div>
          <button onclick={() => restore(p.id)} class="btn btn-ghost btn-xs opacity-0 group-hover:opacity-100">Restore</button>
          <button onclick={() => deleteConfirmId = p.id} class="btn btn-ghost btn-xs text-error opacity-0 group-hover:opacity-100">Delete forever</button>
        </div>
      {/each}
    </div>
  {/if}
</div>
