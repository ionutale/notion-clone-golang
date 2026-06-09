<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { goto } from '$app/navigation';

  let { blockId }:
    {
      blockId: string;
    } = $props();

  let block = $derived(blockStore.blocks.get(blockId));
  let title = $derived(block?.content?.title ?? 'Untitled');

  async function navigate() {
    await goto(`/pages/${blockId}`);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      navigate();
    }
  }
</script>

<div
  class="page-block flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-base-200/70 cursor-pointer transition-colors group"
  onclick={navigate}
  onkeydown={handleKeydown}
  tabindex="0"
  role="button"
>
  {#if block?.content?.icon}
    {#if block.content.icon_type === 'image'}
      <img src={block.content.icon} alt="" class="w-4 h-4 rounded object-cover shrink-0" />
    {:else}
      <span class="text-sm shrink-0">{block.content.icon}</span>
    {/if}
  {:else}
    <svg class="w-4 h-4 text-base-content/40 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
    </svg>
  {/if}
  <span class="text-sm">{title}</span>
</div>
