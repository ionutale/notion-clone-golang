<script lang="ts">
  import type { BlockType } from '$lib/types';
  import { blockStore } from '$lib/stores/blocks.svelte';

  let { position, parentBlockId, onClose, mode = 'create', blockId }: {
    position: { x: number; y: number };
    parentBlockId: string;
    onClose: () => void;
    mode?: 'create' | 'transform';
    blockId?: string;
  } = $props();

  let items: { type: BlockType; label: string; icon: string }[] = [
    { type: 'text', label: 'Text', icon: 'T' },
    { type: 'heading_1', label: 'Heading 1', icon: 'H1' },
    { type: 'heading_2', label: 'Heading 2', icon: 'H2' },
    { type: 'heading_3', label: 'Heading 3', icon: 'H3' },
    { type: 'bullet_list_item', label: 'Bullet List', icon: '•' },
    { type: 'numbered_list_item', label: 'Numbered List', icon: '1.' },
    { type: 'toggle', label: 'Toggle', icon: '▶' },
    { type: 'divider', label: 'Divider', icon: '—' },
    { type: 'image', label: 'Image', icon: '🖼' },
    { type: 'page', label: 'Page', icon: '📄' },
  ];

  let filter = $state('');
  let selectedIndex = $state(0);
  let el = $state<HTMLDivElement>();

  let filtered = $derived(
    filter ? items.filter(i => i.label.toLowerCase().includes(filter.toLowerCase())) : items
  );

  $effect(() => {
    selectedIndex = 0;
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, filtered.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, 0);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (filtered[selectedIndex]) {
        select(filtered[selectedIndex].type);
      }
    } else if (e.key === 'Backspace' && filter === '') {
      e.preventDefault();
      onClose();
    } else if (e.key.length === 1) {
      filter += e.key;
    }
  }

  async function select(type: BlockType) {
    if (mode === 'transform' && blockId) {
      await blockStore.updateBlock(blockId, { type, content: { html: '' } });
    } else {
      const content = type === 'page' ? { title: 'New Page' } : { html: '' };
      await blockStore.createBlock(parentBlockId, type, content);
    }
    onClose();
  }

  function handleClickOutside(e: MouseEvent) {
    if (el && !el.contains(e.target as Node)) {
      onClose();
    }
  }

  $effect(() => {
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  });
</script>

<svelte:window onkeydown={handleKeydown} />

<div
  bind:this={el}
  class="fixed z-50 w-64 bg-base-100 border border-base-300 rounded-xl shadow-xl overflow-hidden"
  style="left: {position.x}px; top: {position.y}px;"
  role="listbox"
>
  <div class="p-2 border-b border-base-200">
    <input
      type="text"
      placeholder="Filter..."
      bind:value={filter}
      autofocus
      class="input input-ghost input-xs w-full"
    />
  </div>
  <div class="max-h-64 overflow-y-auto py-1">
    {#each filtered as item, i (item.type)}
      <button
        class={['w-full text-left px-3 py-2 flex items-center gap-3 text-sm transition-colors', { 'bg-base-200': i === selectedIndex, 'hover:bg-base-200': i !== selectedIndex }]}
        onmouseenter={() => selectedIndex = i}
        onmousedown={() => select(item.type)}
        role="option"
        aria-selected={i === selectedIndex}
      >
        <span class="w-6 h-6 flex items-center justify-center text-xs font-medium bg-base-200 rounded shrink-0">
          {item.icon}
        </span>
        <span>{item.label}</span>
      </button>
    {/each}
    {#if filtered.length === 0}
      <div class="px-3 py-4 text-center text-sm text-base-content/40">No results</div>
    {/if}
  </div>
</div>
