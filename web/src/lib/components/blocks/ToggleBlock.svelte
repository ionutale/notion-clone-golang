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

  import { onMount } from 'svelte';

  onMount(() => {
    if (el && shouldFocus) {
      el.focus();
    }
  });

  function toggle() { open = !open; }

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

<div class="toggle-block">
  <div class="flex items-start">
    <button onclick={toggle} class="toggle-btn pt-1 px-1 cursor-pointer select-none text-base-content/40 hover:text-base-content/70 transition-colors">
      {#if open}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
      {:else}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
      {/if}
    </button>
    <div
      bind:this={el}
      contenteditable="true"
      class="block-editor flex-1 text-base-content min-h-[1.5em] outline-none px-1 py-0.5 rounded hover:bg-base-200/50 focus:bg-base-200/50 transition-colors"
      class:is-empty={isEmpty}
      tabindex="0"
      onkeydown={handleKeydown}
      onblur={handleBlur}
      role="textbox"
      aria-multiline="true"
    >
      {block?.content?.html ?? ''}
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
  .block-editor:empty::before,
  .block-editor.is-empty::before {
    content: 'Toggle';
    color: hsl(var(--bc) / 0.3);
    pointer-events: none;
  }
</style>
