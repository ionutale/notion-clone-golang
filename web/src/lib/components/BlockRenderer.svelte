<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { showUndoToast } from '$lib/stores/toast.svelte';
  import TextBlock from './blocks/TextBlock.svelte';
  import HeadingBlock from './blocks/HeadingBlock.svelte';
  import BulletListBlock from './blocks/BulletListBlock.svelte';
  import NumberedListBlock from './blocks/NumberedListBlock.svelte';
  import ToggleBlock from './blocks/ToggleBlock.svelte';
  import DividerBlock from './blocks/DividerBlock.svelte';
  import ImageBlock from './blocks/ImageBlock.svelte';
  import PageBlock from './blocks/PageBlock.svelte';
  import BlockDragHandle from './BlockDragHandle.svelte';
  import BlockRenderer from './BlockRenderer.svelte';

  let { blockId, depth = 0, listIndex = 0, onSlash, focusBlockId }:
    {
      blockId: string;
      depth?: number;
      listIndex?: number;
      onSlash?: (detail: { blockId: string; position: { x: number; y: number }; isTransform?: boolean }) => void;
      focusBlockId?: string | null;
    } = $props();

  let block = $derived(blockStore.blocks.get(blockId));
  let children = $derived(blockStore.childrenMap.get(blockId) ?? []);
  let hovered = $state(false);
  let dragOver = $state(false);

  function handleEnter() {
    createBelow();
  }

  function handleBackspace() {
    deleteBlock();
  }

  async function createBelow(type: string = 'text') {
    const parentId = block?.parent_id ?? null;
    const pos = (block?.position ?? 0) + 1;
    await blockStore.createBlock(parentId, type as any, { html: '' }, pos);
  }

  async function deleteBlock() {
    const deleted = await blockStore.deleteBlock(blockId);
    showUndoToast(deleted);
  }

  async function handleMoveUp() {
    const parentId = block?.parent_id ?? null;
    const pos = Math.max(0, (block?.position ?? 0) - 1);
    if (pos !== block?.position) {
      await blockStore.moveBlock(blockId, parentId, pos);
    }
  }

  async function handleMoveDown() {
    const parentId = block?.parent_id ?? null;
    const pos = (block?.position ?? 0) + 1;
    await blockStore.moveBlock(blockId, parentId, pos);
  }

  function handleSlash(pos: { x: number; y: number }) {
    onSlash?.({ blockId, position: pos, isTransform: true });
  }

  async function handleIndent() {
    if (!block) return;
    const siblings = blockStore.childrenMap.get(block.parent_id) ?? [];
    const idx = siblings.indexOf(blockId);
    if (idx <= 0) return;
    const prevId = siblings[idx - 1];
    const prev = blockStore.blocks.get(prevId);
    if (!prev) return;
    await blockStore.moveBlock(blockId, prevId, 0);
  }

  async function handleOutdent() {
    if (!block || !block.parent_id) return;
    const parent = blockStore.blocks.get(block.parent_id);
    if (!parent) return;
    await blockStore.moveBlock(blockId, parent.parent_id ?? null, (block.position ?? 0) + 1);
  }

  function handleDragStart(e: DragEvent) {
    e.dataTransfer?.setData('text/plain', blockId);
    e.dataTransfer!.effectAllowed = 'move';
    (e.target as HTMLElement)?.closest?.('.block-wrapper')?.classList.add('drag-opacity');
  }

  function handleDragEnd() {
    document.querySelectorAll('.drag-opacity').forEach(el => el.classList.remove('drag-opacity'));
    dragOver = false;
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    e.dataTransfer!.dropEffect = 'move';
    dragOver = true;
  }

  function handleDragLeave() {
    dragOver = false;
  }

  async function handleDrop(e: DragEvent) {
    e.preventDefault();
    dragOver = false;
    const draggedId = e.dataTransfer?.getData('text/plain');
    if (!draggedId || draggedId === blockId) return;
    const parentId = block?.parent_id ?? null;
    await blockStore.moveBlock(draggedId, parentId, block?.position ?? 0);
  }

  // Touch drag support
  let touchDraggedId = $state<string | null>(null);

  function handleTouchStart(_e: TouchEvent) {
    const el = document.querySelector(`[data-block-id="${blockId}"]`);
    el?.classList.add('drag-opacity');
    touchDraggedId = blockId;
  }

  function handleTouchMove(e: TouchEvent) {
    if (!touchDraggedId) return;
    e.preventDefault();
    const x = e.touches[0].clientX;
    const y = e.touches[0].clientY;
    // Remove old indicator
    document.querySelectorAll('.drop-indicator').forEach(el => el.remove());
    // Find block under finger
    const target = document.elementFromPoint(x, y)?.closest('[data-block-id]') as HTMLElement | null;
    if (target && target.getAttribute('data-block-id') !== touchDraggedId) {
      target.insertAdjacentHTML('afterend', '<div class="drop-indicator h-0.5 bg-primary rounded-full mx-1"></div>');
    }
  }

  function handleTouchEnd(_e: TouchEvent) {
    if (!touchDraggedId) return;
    const draggedId = touchDraggedId;
    touchDraggedId = null;
    document.querySelectorAll('.drag-opacity').forEach(el => el.classList.remove('drag-opacity'));
    const indicator = document.querySelector('.drop-indicator');
    if (indicator) {
      const targetWrapper = indicator.previousElementSibling as HTMLElement | null;
      indicator.remove();
      if (targetWrapper) {
        const targetId = targetWrapper.getAttribute('data-block-id');
        if (targetId && targetId !== draggedId) {
          const targetData = blockStore.blocks.get(targetId);
          if (targetData) {
            blockStore.moveBlock(draggedId, targetData.parent_id ?? null, targetData.position ?? 0);
          }
        }
      }
    }
  }

  let needFocus = $derived(focusBlockId === blockId);
</script>

{#if block}
  <div
    class="block-wrapper group relative"
    class:drag-over={dragOver}
    onmouseenter={() => hovered = true}
    onmouseleave={() => hovered = false}
    ondragover={handleDragOver}
    ondragleave={handleDragLeave}
    ondrop={handleDrop}
    ondragend={handleDragEnd}
    ontouchcancel={handleTouchEnd}
    role="listitem"
    data-block-id={blockId}
  >
    <div class="flex items-start gap-0.5" style="margin-left: {depth * 1.5}rem;">
      <div class="flex items-center h-8 w-6 shrink-0 -ml-6 opacity-0 group-hover:opacity-100 transition-opacity">
        <BlockDragHandle
          {blockId}
          onDragStart={handleDragStart}
          visible={hovered}
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
        />
      </div>
      <div class="flex-1 min-w-0">
        {#if block.type === 'text'}
          <TextBlock {blockId} onEnter={handleEnter} onBackspace={handleBackspace} onSlash={handleSlash} onMoveUp={handleMoveUp} onMoveDown={handleMoveDown} onIndent={handleIndent} onOutdent={handleOutdent} shouldFocus={needFocus} />
        {:else if block.type === 'heading_1' || block.type === 'heading_2' || block.type === 'heading_3'}
          <HeadingBlock {blockId} onEnter={handleEnter} onBackspace={handleBackspace} onSlash={handleSlash} onMoveUp={handleMoveUp} onMoveDown={handleMoveDown} onIndent={handleIndent} onOutdent={handleOutdent} shouldFocus={needFocus} />
        {:else if block.type === 'bullet_list_item'}
          <BulletListBlock {blockId} onEnter={handleEnter} onBackspace={handleBackspace} onSlash={handleSlash} onMoveUp={handleMoveUp} onMoveDown={handleMoveDown} onIndent={handleIndent} onOutdent={handleOutdent} shouldFocus={needFocus} />
        {:else if block.type === 'numbered_list_item'}
          <NumberedListBlock {blockId} index={listIndex} onEnter={handleEnter} onBackspace={handleBackspace} onSlash={handleSlash} onMoveUp={handleMoveUp} onMoveDown={handleMoveDown} onIndent={handleIndent} onOutdent={handleOutdent} shouldFocus={needFocus} />
        {:else if block.type === 'toggle'}
          <ToggleBlock {blockId} onEnter={handleEnter} onBackspace={handleBackspace} onSlash={handleSlash} onMoveUp={handleMoveUp} onMoveDown={handleMoveDown} onIndent={handleIndent} onOutdent={handleOutdent} shouldFocus={needFocus} />
        {:else if block.type === 'divider'}
          <DividerBlock onEnter={handleEnter} onBackspace={handleBackspace} onMoveUp={handleMoveUp} onMoveDown={handleMoveDown} onIndent={handleIndent} onOutdent={handleOutdent} />
        {:else if block.type === 'image'}
          <ImageBlock {blockId} onEnter={handleEnter} onBackspace={handleBackspace} onMoveUp={handleMoveUp} onMoveDown={handleMoveDown} onIndent={handleIndent} onOutdent={handleOutdent} />
        {:else if block.type === 'page'}
          <PageBlock {blockId} />
        {/if}

        {#if children.length > 0}
          <div class="space-y-0.5">
            {#each children as childId, i (childId)}
              <BlockRenderer
                blockId={childId}
                depth={depth + 1}
                listIndex={i + 1}
                {onSlash}
                {focusBlockId}
              />
            {/each}
          </div>
        {/if}
      </div>
    </div>

    {#if block.type !== 'divider'}
      <button
        onclick={deleteBlock}
        class="absolute right-0 top-0.5 btn btn-ghost btn-xs px-1 opacity-0 group-hover:opacity-100 transition-opacity text-base-content/30 hover:text-error"
        title="Delete block"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    {/if}
  </div>

  {#if dragOver}
    <div class="h-0.5 bg-primary rounded-full mx-1 transition-all"></div>
  {/if}
{/if}

<style>
  .drag-opacity {
    opacity: 0.4;
  }
</style>
