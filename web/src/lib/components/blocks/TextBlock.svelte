<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { api } from '$lib/api';

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

  function handleInput() {
    if (!el) return;
  }

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
    } else if (e.key === '/' && (el?.textContent ?? '') === '') {
      e.preventDefault();
      e.stopPropagation();
      if (!el) return;
      const rect = el.getBoundingClientRect();
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

  async function save() {
    if (!el) return;
    await blockStore.updateBlock(blockId, { content: { html: el.innerHTML } });
  }

  function handleBlur() {
    focused = false;
    save();
  }

  async function handlePaste(e: ClipboardEvent) {
    const html = e.clipboardData?.getData('text/html');
    if (html && html.includes('<img')) {
      e.preventDefault();
      const match = html.match(/<img[^>]+src=["']([^"']+)["']/);
      const url = match ? match[1] : null;
      if (url) {
        const block = blockStore.blocks.get(blockId);
        if (block) {
          await blockStore.createBlock(block.parent_id ?? blockId, 'image', { url }, (block.position ?? 0) + 1);
        }
      }
      return;
    }
    const files = e.clipboardData?.files;
    if (files && files.length > 0 && files[0].type.startsWith('image/')) {
      e.preventDefault();
      try {
        const { url } = await api.uploadFile(files[0]);
        const block = blockStore.blocks.get(blockId);
        if (block) {
          await blockStore.createBlock(block.parent_id ?? blockId, 'image', { url }, (block.position ?? 0) + 1);
        }
      } catch {
        // silent
      }
    }
  }

  let isEmpty = $derived(!block?.content?.html || block.content.html === '<br>');
  let focused = $state(false);
  let showPlaceholder = $derived(isEmpty && !focused);
</script>

<div class="relative">
  {#if showPlaceholder}
    <span class="absolute left-1 top-0.5 pointer-events-none" style="color: hsl(var(--bc) / 0.3); line-height: 1.5rem;">Type / for commands</span>
  {/if}
  <div
    bind:this={el}
    contenteditable="true"
    class={['block-editor text-base-content min-h-[1.5em] outline-none px-1 py-0.5 rounded hover:bg-base-200/50 focus:bg-base-200/50 transition-colors', { 'is-empty': isEmpty }]}
    tabindex="0"
    onfocus={() => focused = true}
    oninput={handleInput}
    onkeydown={handleKeydown}
    onblur={handleBlur}
    onpaste={handlePaste}
    role="textbox"
    aria-multiline="true"
  ></div>
</div>

<style>
  .block-editor :global(a) {
    color: hsl(var(--p));
    text-decoration: underline;
    cursor: pointer;
  }
</style>
