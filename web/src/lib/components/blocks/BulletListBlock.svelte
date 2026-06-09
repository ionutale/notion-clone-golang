<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';

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

  $effect(() => {
    const html = block?.content?.html ?? '';
    if (el && el.innerHTML !== html) {
      el.innerHTML = html;
    }
  });

  $effect(() => {
    if (el && shouldFocus) {
      el.focus();
    }
  });

  async function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      e.stopPropagation();
      if (el && (el.textContent ?? '').trim() === '') return;
      await save();
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
    } else if (e.key === '/' && (el?.textContent ?? '') === '') {
      e.preventDefault();
      e.stopPropagation();
      if (!el) return;
      const rect = el.getBoundingClientRect();
      onSlash({ x: rect.left, y: rect.bottom });
      el.textContent = '';
    } else if (e.key === 'Tab' && !e.shiftKey) {
      e.preventDefault();
      onIndent?.();
    } else if (e.key === 'Tab' && e.shiftKey) {
      e.preventDefault();
      onOutdent?.();
    }
  }

  async function save() {
    if (!el) return;
    await blockStore.updateBlock(blockId, { content: { html: el.innerHTML } });
  }

  function handleBlur() {
    focused = false;
    save();
  }

  let isEmpty = $derived(!block?.content?.html || block.content.html === '<br>');
  let focused = $state(false);
  let showPlaceholder = $derived(isEmpty && !focused);
</script>

<li class="list-item">
  <span class="bullet-marker text-base-content/40 select-none mr-2">•</span>
  <div class="relative flex-1">
    {#if showPlaceholder}
      <span class="absolute left-0 top-0 pointer-events-none" style="color: hsl(var(--bc) / 0.3); line-height: 1.5rem;">List item</span>
    {/if}
    <div
      bind:this={el}
      contenteditable="true"
      class={['block-editor text-base-content min-h-[1.5em] outline-none', { 'is-empty': isEmpty }]}
      tabindex="0"
      onfocus={() => focused = true}
      onkeydown={handleKeydown}
      onblur={handleBlur}
    role="textbox"
    aria-multiline="true"
  ></div>
  </div>
</li>

<style>
  .list-item {
    display: flex;
    align-items: flex-start;
    padding: 0.125rem 0;
  }
  .bullet-marker {
    line-height: 1.5rem;
    font-size: 1.25rem;
  }
  .block-editor :global(a) {
    color: hsl(var(--p));
    text-decoration: underline;
    cursor: pointer;
  }
</style>
