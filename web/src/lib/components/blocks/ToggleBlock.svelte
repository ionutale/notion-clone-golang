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
  let open = $state(false);
  let children = $derived(blockStore.childrenMap.get(blockId) ?? []);

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

  function toggle() { open = !open; }

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

<div class="toggle-block">
  <div class="flex items-start">
    <button onclick={toggle} class="toggle-btn pt-1 px-1 cursor-pointer select-none text-base-content/40 hover:text-base-content/70 transition-colors">
      {#if open}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
      {:else}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
      {/if}
    </button>
    <div class="relative flex-1">
      {#if showPlaceholder}
        <span class="absolute left-1 top-0.5 pointer-events-none" style="color: hsl(var(--bc) / 0.3); line-height: 1.5rem;">Toggle</span>
      {/if}
      <div
        bind:this={el}
        contenteditable="true"
        class={['block-editor text-base-content min-h-[1.5em] outline-none px-1 py-0.5 rounded hover:bg-base-200/50 focus:bg-base-200/50 transition-colors', { 'is-empty': isEmpty }]}
        tabindex="0"
        onfocus={() => focused = true}
        onkeydown={handleKeydown}
        onblur={handleBlur}
    role="textbox"
    aria-multiline="true"
  ></div>
    </div>
  </div>
  {#if open && children.length > 0}
    <div class="ml-6 space-y-0.5">
      {#each children as childId (childId)}
        <div>child block {childId}</div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .toggle-btn {
    background: none;
    border: none;
    outline: none;
  }
  .block-editor :global(a) {
    color: hsl(var(--p));
    text-decoration: underline;
    cursor: pointer;
  }
</style>
