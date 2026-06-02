<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import type { BlockType } from '$lib/types';

  let { blockId, onEnter, onBackspace, onSlash, onMoveUp, onMoveDown, onIndent, onOutdent, shouldFocus = false }:
    {
      blockId: string;
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

  let headingTag = $derived(block?.type === 'heading_1' ? 'h1' : block?.type === 'heading_2' ? 'h2' : 'h3');
  let headingSize = $derived(
    block?.type === 'heading_1' ? 'text-3xl' : block?.type === 'heading_2' ? 'text-2xl' : 'text-xl'
  );
  let headingWeight = $derived(block?.type === 'heading_1' ? 'font-bold' : block?.type === 'heading_2' ? 'font-semibold' : 'font-semibold');

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

<svelte:element this={headingTag}
  bind:this={el}
  contenteditable="true"
  class="block-editor outline-none px-1 py-0.5 rounded hover:bg-base-200/50 focus:bg-base-200/50 transition-colors {headingSize} {headingWeight}"
  class:is-empty={isEmpty}
  tabindex="0"
  onkeydown={handleKeydown}
  onblur={handleBlur}
  role="textbox"
  aria-multiline="true"
>
  {block?.content?.html ?? ''}
</svelte:element>

<style>
  .block-editor:empty::before,
  .block-editor.is-empty::before {
    content: 'Untitled';
    color: hsl(var(--bc) / 0.3);
    pointer-events: none;
  }
</style>
