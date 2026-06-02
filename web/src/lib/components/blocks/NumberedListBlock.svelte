<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';

  let { blockId, index, onEnter, onBackspace, onSlash, onMoveUp, onMoveDown, onIndent, onOutdent, shouldFocus = false }:
    {
      blockId: string;
      index: number;
      onEnter: () => void;
      onBackspace: () => void;
      onSlash: (pos: { x: number; y: number }) => void;
      onMoveUp: () => void;
      onMoveDown: () => void;
      onIndent?: () => void;
      onOutdent?: () => void;
      shouldFocus?: boolean;
    } = $props();

  let block = $derived(blockStore.blocks.get(blockId));
  let el = $state<HTMLDivElement>();

  import { onMount } from 'svelte';

  onMount(() => {
    if (el && shouldFocus) {
      el.focus();
    }
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      save();
      onEnter();
    } else if (e.key === 'Backspace') {
      if (el && (el.textContent ?? '').trim() === '') {
        e.preventDefault();
        onBackspace();
      }
    } else if (e.altKey && e.key === 'ArrowUp') {
      e.preventDefault();
      onMoveUp();
    } else if (e.altKey && e.key === 'ArrowDown') {
      e.preventDefault();
      onMoveDown();
    } else if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
      e.preventDefault();
      document.execCommand('bold');
    } else if ((e.metaKey || e.ctrlKey) && e.key === 'i') {
      e.preventDefault();
      document.execCommand('italic');
    } else if ((e.metaKey || e.ctrlKey) && e.key === 'u') {
      e.preventDefault();
      document.execCommand('underline');
    } else if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'S') {
      e.preventDefault();
      document.execCommand('strikeThrough');
    } else if (e.key === 'Tab' && !e.shiftKey) {
      e.preventDefault();
      onIndent?.();
    } else if (e.key === 'Tab' && e.shiftKey) {
      e.preventDefault();
      onOutdent?.();
    }
  }

  function save() {
    if (!el) return;
    blockStore.updateBlock(blockId, { content: { html: el.innerHTML } });
  }

  function handleBlur() {
    save();
  }

  let isEmpty = $derived(!block?.content?.html || block.content.html === '<br>');
</script>

<li class="list-item" style="list-style: none;">
  <span class="number-marker text-base-content/40 select-none mr-2 tabular-nums">{index}.</span>
  <div
    bind:this={el}
    contenteditable="true"
    class="block-editor flex-1 text-base-content min-h-[1.5em] outline-none"
    class:is-empty={isEmpty}
    tabindex="0"
    onkeydown={handleKeydown}
    onblur={handleBlur}
    role="textbox"
    aria-multiline="true"
  >
    {block?.content?.html ?? ''}
  </div>
</li>

<style>
  .list-item {
    display: flex;
    align-items: flex-start;
    padding: 0.125rem 0;
  }
  .number-marker {
    line-height: 1.5rem;
    min-width: 1.5rem;
    text-align: right;
  }
  .block-editor:empty::before,
  .block-editor.is-empty::before {
    content: 'List item';
    color: hsl(var(--bc) / 0.3);
    pointer-events: none;
  }
</style>
