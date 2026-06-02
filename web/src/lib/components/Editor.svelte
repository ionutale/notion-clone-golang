<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import BlockRenderer from './BlockRenderer.svelte';
  import SlashMenu from './SlashMenu.svelte';
  import UndoToast from './UndoToast.svelte';
  import { page } from '$app/stores';

  let { pageId } = $props<{ pageId: string }>();

  let slashMenu = $state<{ blockId: string; position: { x: number; y: number }; isTransform: boolean } | null>(null);
  let focusBlockId = $state<string | null>(null);
  let undoStack = $state<Array<() => void>>([]);

  $effect(() => {
    if (pageId) {
      blockStore.loadPage(pageId);
      focusBlockId = null;
    }
    return () => {
      blockStore.clear();
    };
  });

  $effect(() => {
    // Update document title
    if (blockStore.pageTitle) {
      document.title = `${blockStore.pageTitle} - Notion Clone`;
    }
  });

  function handleSlash(detail: { blockId: string; position: { x: number; y: number }; isTransform?: boolean }) {
    slashMenu = { ...detail, isTransform: detail.isTransform ?? true };
  }

  function closeSlashMenu() {
    slashMenu = null;
  }

  async function handlePageKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      slashMenu = null;
    } else if ((e.metaKey || e.ctrlKey) && e.key === 'z') {
      e.preventDefault();
      const undo = undoStack.pop();
      if (undo) undo();
    }
  }

  function pushUndo(fn: () => void) {
    undoStack = [...undoStack, fn];
  }

  async function addBlockAtBottom() {
    const block = await blockStore.createBlock(null, 'text', { html: '' });
    focusBlockId = block.id;
  }
</script>

<svelte:window onkeydown={handlePageKeydown} />

<div class="max-w-3xl mx-auto py-8 px-4">
  {#if blockStore.loading}
    <div class="flex justify-center py-20">
      <span class="loading loading-spinner loading-lg text-primary"></span>
    </div>
  {:else if blockStore.error}
    <div class="alert alert-error shadow-lg my-8">
      <svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01M12 2a10 10 0 100 20 10 10 0 000-20z" />
      </svg>
      <span>{blockStore.error}</span>
    </div>
  {:else}
    <!-- Page Title -->
    <h1
      class="text-4xl font-bold mb-8 text-base-content outline-none"
    >
      {blockStore.pageTitle}
    </h1>

    <!-- Blocks -->
    <div class="space-y-0.5" role="list">
      {#each blockStore.rootBlocks as blockId (blockId)}
        <BlockRenderer
          {blockId}
          {focusBlockId}
          onSlash={handleSlash}
        />
      {/each}
    </div>

    <!-- Bottom new-block trigger -->
    <button
      onclick={addBlockAtBottom}
      class="mt-2 w-full text-left px-1 py-2 text-sm text-base-content/30 hover:text-base-content/50 transition-colors rounded hover:bg-base-200/50"
    >
      <svg class="w-4 h-4 inline mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
      Click to add a block
    </button>
  {/if}
</div>

{#if slashMenu}
  <SlashMenu
    position={slashMenu.position}
    parentBlockId={slashMenu.blockId}
    onClose={closeSlashMenu}
    mode={slashMenu.isTransform ? 'transform' : 'create'}
    blockId={slashMenu.isTransform ? slashMenu.blockId : undefined}
  />
{/if}

<UndoToast />
