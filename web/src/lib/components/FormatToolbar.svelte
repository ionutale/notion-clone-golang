<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import PromptDialog from '$lib/components/PromptDialog.svelte';
  import { execFormat, queryFormatState } from '$lib/format';

  let activeFormats = $state({
    bold: false,
    italic: false,
    underline: false,
    strikeThrough: false,
  });

  let headingOpen = $state(false);
  let headingContainer = $state<HTMLDivElement>();
  let showLinkPrompt = $state(false);

  $effect(() => {
    if (!headingOpen) return;
    const handler = (e: MouseEvent) => {
      if (headingContainer && !headingContainer.contains(e.target as Node)) {
        headingOpen = false;
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  });

  function exec(cmd: string, val?: string) {
    execFormat(cmd, val);
    document.getSelection()?.getRangeAt(0)?.collapse(true);
    updateActiveState();
  }

  function updateActiveState() {
    activeFormats = {
      bold: queryFormatState('bold'),
      italic: queryFormatState('italic'),
      underline: queryFormatState('underline'),
      strikeThrough: queryFormatState('strikeThrough'),
    };
  }

  $effect(() => {
    const handler = () => updateActiveState();
    document.addEventListener('selectionchange', handler);
    return () => document.removeEventListener('selectionchange', handler);
  });

  function handleLink() {
    const sel = window.getSelection();
    if (!sel || sel.isCollapsed) return;
    showLinkPrompt = true;
  }

  function applyLink(url: string) {
    if (url) {
      execFormat('createLink', url);
    }
    showLinkPrompt = false;
  }

  function handleCode() {
    const sel = window.getSelection();
    if (!sel || !sel.rangeCount) return;
    const range = sel.getRangeAt(0);
    if (range.collapsed) return;
    const parent = range.commonAncestorContainer.parentElement;
    if (parent?.tagName === 'CODE') {
      const text = parent.textContent || '';
      parent.replaceWith(text);
      sel.removeAllRanges();
    } else {
      execFormat('insertHTML', `<code>${range.toString()}</code>`);
    }
  }

  function handleHeading(type: string) {
    const blockId = findActiveBlock();
    if (blockId) {
      blockStore.updateBlock(blockId, { type: type as any });
    }
    headingOpen = false;
  }

  function handleClearFormatting() {
    execFormat('removeFormat');
  }

  function findActiveBlock(): string | null {
    const sel = window.getSelection();
    if (!sel || !sel.rangeCount) return null;
    let el: Node | null = sel.getRangeAt(0).commonAncestorContainer;
    while (el && !(el instanceof HTMLElement)) {
      el = el.parentElement;
    }
    let htmlEl = el as HTMLElement | null;
    while (htmlEl) {
      if (htmlEl.getAttribute('contenteditable') === 'true') {
        const wrapper = htmlEl.closest('[data-block-id]');
        return wrapper?.getAttribute('data-block-id') ?? null;
      }
      htmlEl = htmlEl.parentElement;
    }
    return null;
  }
</script>

<PromptDialog
  open={showLinkPrompt}
  title="Enter URL"
  placeholder="https://example.com"
  confirmText="Apply"
  onConfirm={applyLink}
  onCancel={() => showLinkPrompt = false}
/>

<div class="flex flex-wrap items-center gap-1 px-3 py-2 mb-4 bg-base-200/70 rounded-xl border border-base-300 shadow-sm" role="toolbar" aria-label="Text formatting" tabindex="0" onmousedown={(e) => e.preventDefault()}>
  <!-- Row 1: Inline Formatting -->
  <div class="flex items-center gap-1">
    <button
      onclick={() => exec('bold')}
      class={['btn btn-ghost btn-xs px-2', { 'btn-active': activeFormats.bold }]}
      title="Bold (Cmd+B)"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 4h6a4 4 0 014 4 4 4 0 01-4 4H6zM6 12h8a4 4 0 014 4 4 4 0 01-4 4H6z"/></svg>
    </button>
    <button
      onclick={() => exec('italic')}
      class={['btn btn-ghost btn-xs px-2', { 'btn-active': activeFormats.italic }]}
      title="Italic (Cmd+I)"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 4h8M6 20h8M14 4l-4 16"/></svg>
    </button>
    <button
      onclick={() => exec('underline')}
      class={['btn btn-ghost btn-xs px-2', { 'btn-active': activeFormats.underline }]}
      title="Underline (Cmd+U)"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 20h12M6 4v6a6 6 0 0012 0V4"/></svg>
    </button>
    <button
      onclick={() => exec('strikeThrough')}
      class={['btn btn-ghost btn-xs px-2', { 'btn-active': activeFormats.strikeThrough }]}
      title="Strikethrough (Cmd+Shift+S)"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 12h12M8 7V4h8v3M8 17v3h8v-3"/></svg>
    </button>
    <button
      onclick={handleCode}
      class="btn btn-ghost btn-xs px-2"
      title="Inline Code"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/></svg>
    </button>
  </div>

  <div class="w-px h-5 bg-base-300 mx-1"></div>

  <!-- Row 2: Block Actions -->
  <div class="flex items-center gap-1">
    <div bind:this={headingContainer} class="relative">
      <button
        onclick={() => headingOpen = !headingOpen}
        class="btn btn-ghost btn-xs px-2"
        title="Heading"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h8"/></svg>
      </button>
      {#if headingOpen}
        <div class="absolute top-full left-0 mt-1 w-36 bg-base-100 border border-base-300 rounded-lg shadow-xl z-50 py-1">
          <button onclick={() => handleHeading('heading_1')} title="Heading 1" class="w-full text-left px-3 py-1.5 text-sm hover:bg-base-200 text-lg font-bold">H1</button>
          <button onclick={() => handleHeading('heading_2')} title="Heading 2" class="w-full text-left px-3 py-1.5 text-sm hover:bg-base-200 text-base font-semibold">H2</button>
          <button onclick={() => handleHeading('heading_3')} title="Heading 3" class="w-full text-left px-3 py-1.5 text-sm hover:bg-base-200 text-sm font-semibold">H3</button>
        </div>
      {/if}
    </div>
    <button
      onclick={handleLink}
      class="btn btn-ghost btn-xs px-2"
      title="Link"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/></svg>
    </button>
    <button
      onclick={handleClearFormatting}
      class="btn btn-ghost btn-xs px-2"
      title="Clear Formatting"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
    </button>
  </div>
</div>