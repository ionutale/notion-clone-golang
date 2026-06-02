<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { api } from '$lib/api';

  let { blockId, onEnter, onBackspace, onMoveUp, onMoveDown, onIndent, onOutdent }:
    {
      blockId: string;
      onEnter: () => void;
      onBackspace: () => void;
      onMoveUp: () => void;
      onMoveDown: () => void;
      onIndent?: () => void;
      onOutdent?: () => void;
    } = $props();

  let block = $derived(blockStore.blocks.get(blockId));
  let uploading = $state(false);
  let fileInput = $state<HTMLInputElement>();

  async function handleFile(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    uploading = true;
    try {
      const { url } = await api.uploadFile(file);
      await blockStore.updateBlock(blockId, { content: { url } });
    } catch (err) {
      console.error('Upload failed', err);
    } finally {
      uploading = false;
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      onEnter();
    } else if (e.key === 'Backspace') {
      if (!block?.content?.url) {
        e.preventDefault();
        onBackspace();
      }
    } else if (e.altKey && e.key === 'ArrowUp') {
      e.preventDefault();
      onMoveUp();
    } else if (e.altKey && e.key === 'ArrowDown') {
      e.preventDefault();
      onMoveDown();
    } else if (e.key === 'Tab' && !e.shiftKey) {
      e.preventDefault();
      onIndent?.();
    } else if (e.key === 'Tab' && e.shiftKey) {
      e.preventDefault();
      onOutdent?.();
    }
  }

  function openFilePicker() {
    fileInput?.click();
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex a11y_no_noninteractive_element_interactions -->
<div
  class="image-block group relative my-1 px-1"
  tabindex="0"
  onkeydown={handleKeydown}
  aria-label="Image block"
>
  {#if block?.content?.url}
    <figure class="relative">
      <img src={block.content.url} alt="" class="max-w-full rounded-lg" />
      <button
        onclick={() => blockStore.updateBlock(blockId, { content: {} })}
        class="absolute top-2 right-2 btn btn-ghost btn-xs opacity-0 group-hover:opacity-100 transition-opacity"
      >
        Remove
      </button>
    </figure>
  {:else if uploading}
    <div class="flex items-center gap-2 p-4 text-base-content/50">
      <span class="loading loading-spinner loading-sm"></span>
      Uploading...
    </div>
  {:else}
    <button
      onclick={openFilePicker}
      class="w-full p-8 border-2 border-dashed border-base-300 rounded-lg text-base-content/40 hover:text-base-content/60 hover:border-base-content/40 transition-colors text-sm"
    >
      Click to upload an image
    </button>
  {/if}
  <input
    bind:this={fileInput}
    type="file"
    accept="image/*"
    class="hidden"
    onchange={handleFile}
  />
</div>
