<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';

  let { blockId, onEnter, onBackspace, onMoveUp, onMoveDown, onIndent, onOutdent, onSlash, shouldFocus = false }:
    {
      blockId: string;
      onEnter: () => void;
      onBackspace: () => void;
      onMoveUp: () => void;
      onMoveDown: () => void;
      onIndent?: () => void;
      onOutdent?: () => void;
      onSlash: (pos: { x: number; y: number }) => void;
      shouldFocus?: boolean;
    } = $props();

  let el = $state<HTMLPreElement>();

  let block = $derived(blockStore.blocks.get(blockId));
  let code = $derived(block?.content?.code ?? '');

  let savedCode = $state('');

  function handleInput() {
    if (!el) return;
    const text = el.textContent ?? '';
    if (text !== savedCode) {
      savedCode = text;
      blockStore.updateBlock(blockId, { content: { code: text } });
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      onEnter();
    } else if ((e.key === 'Backspace' || e.key === 'Delete') && (el?.textContent ?? '') === '') {
      e.preventDefault();
      onBackspace();
    } else if (e.key === 'Tab') {
      e.preventDefault();
      if (e.shiftKey) {
        onOutdent?.();
      } else {
        onIndent?.();
      }
    } else if ((e.metaKey || e.ctrlKey) && e.key === 'b') {
      e.preventDefault();
    } else if ((e.metaKey || e.ctrlKey) && e.key === 'i') {
      e.preventDefault();
    } else if ((e.metaKey || e.ctrlKey) && e.key === 'u') {
      e.preventDefault();
    }
  }

  $effect(() => {
    if (shouldFocus && el) {
      el.focus();
    }
  });

  $effect(() => {
    if (el && block && block.content?.code !== undefined) {
      if (el.textContent !== block.content.code) {
        el.textContent = block.content.code;
      }
    }
  });
</script>

<pre
  bind:this={el}
  contenteditable="true"
  class="code-block bg-base-300/50 border border-base-300 rounded-lg p-4 font-mono text-sm leading-relaxed whitespace-pre-wrap overflow-x-auto min-h-[3rem] focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-colors"
  class:opacity-60={!block}
  oninput={handleInput}
  onkeydown={handleKeydown}
  role="textbox"
  aria-multiline="true"
  aria-label="Code block"
  data-block-id={blockId}
>{code}</pre>
