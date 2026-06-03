<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import BlockRenderer from './BlockRenderer.svelte';
  import FormatToolbar from './FormatToolbar.svelte';
  import SlashMenu from './SlashMenu.svelte';
  import UndoToast from './UndoToast.svelte';
  import IconPopover from './IconPopover.svelte';
  import { api } from '$lib/api';

  let { pageId } = $props<{ pageId: string }>();

  let showIconPicker = $state(false);

  function handleIconPickerOutsideClick(e: MouseEvent) {
    if (showIconPicker) {
      showIconPicker = false;
    }
  }

  let slashMenu = $state<{ blockId: string; position: { x: number; y: number }; isTransform: boolean } | null>(null);
  let focusBlockId = $state<string | null>(null);

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
    }
  }

  async function addBlockAtBottom() {
    const block = await blockStore.createBlock(null, 'text', { html: '' });
    focusBlockId = block.id;
  }
</script>

<svelte:window onkeydown={handlePageKeydown} onclick={handleIconPickerOutsideClick} />

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
    <div class="flex items-start gap-4 mb-8">
      {#if blockStore.pageIcon}
        <button
          onclick={() => showIconPicker = !showIconPicker}
          class="shrink-0 w-12 h-12 flex items-center justify-center text-4xl rounded-xl hover:bg-base-200 transition-colors relative"
        >
          {#if blockStore.pageIconType === 'image'}
            <img src={blockStore.pageIcon} alt="Page icon" class="w-12 h-12 rounded object-cover" />
          {:else}
            {blockStore.pageIcon}
          {/if}
        </button>
      {:else}
        <button
          onclick={() => showIconPicker = !showIconPicker}
          class="shrink-0 w-12 h-12 flex items-center justify-center text-2xl rounded-xl hover:bg-base-200 transition-colors text-base-content/20 hover:text-base-content/40"
        >
          +
        </button>
      {/if}
      <h1 class="text-4xl font-bold text-base-content outline-none flex-1 min-w-0">
        {blockStore.pageTitle}
      </h1>
    </div>

    {#if showIconPicker}
      <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
      <div
        class="relative"
        onclick={(e) => e.stopPropagation()}
      >
        <IconPopover
          onselect={async (detail) => {
            await blockStore.updateIcon(detail.value, detail.type);
            showIconPicker = false;
          }}
          onremove={async () => {
            await blockStore.updateIcon(null, null);
            showIconPicker = false;
          }}
          onclose={() => showIconPicker = false}
        />
      </div>
    {/if}

    <FormatToolbar />

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
