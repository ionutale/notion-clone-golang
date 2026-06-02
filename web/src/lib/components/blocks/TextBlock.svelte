<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { onMount } from 'svelte';

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

  onMount(() => {
    if (el && shouldFocus) {
      el.focus();
    }
  });

  function handleInput() {
    if (!el) return;
  }

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
    } else if (e.key === '/' && (el?.textContent ?? '') === '') {
      e.preventDefault();
      const rect = el!.getBoundingClientRect();
      onSlash({ x: rect.left, y: rect.bottom });
      el.textContent = '';
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
    const html = el.innerHTML;
    blockStore.updateBlock(blockId, { content: { html } });
  }

  function handleBlur() {
    save();
  }

  let isEmpty = $derived(!block?.content?.html || block.content.html === '<br>');
</script>

<div
  bind:this={el}
  contenteditable="true"
  class="block-editor text-base-content min-h-[1.5em] outline-none px-1 py-0.5 rounded hover:bg-base-200/50 focus:bg-base-200/50 transition-colors"
  class:is-empty={isEmpty}
  tabindex="0"
  oninput={handleInput}
  onkeydown={handleKeydown}
  onblur={handleBlur}
  role="textbox"
  aria-multiline="true"
>
  {block?.content?.html ?? ''}
</div>

<style>
  .block-editor:empty::before,
  .block-editor.is-empty::before {
    content: 'Type / for commands';
    color: hsl(var(--bc) / 0.3);
    pointer-events: none;
  }
  .block-editor:focus.is-empty::before {
    content: 'Type / for commands';
  }
</style>
