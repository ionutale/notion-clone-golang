<script lang="ts">
  let { onEnter, onBackspace, onMoveUp, onMoveDown, onIndent, onOutdent }:
    {
      onEnter: () => void;
      onBackspace: () => void;
      onMoveUp: () => void;
      onMoveDown: () => void;
      onIndent?: () => void;
      onOutdent?: () => void;
    } = $props();

  let el = $state<HTMLDivElement>();

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      onEnter();
    } else if (e.key === 'Backspace') {
      e.preventDefault();
      onBackspace();
    } else if (e.altKey && e.key === 'ArrowUp') {
      e.preventDefault();
      onMoveUp();
    } else if (e.altKey && e.key === 'ArrowDown') {
      e.preventDefault();
      onMoveDown();
    } else if (e.key === 'Tab' && !e.shiftKey) {
      e.preventDefault();
      onIndent?.();
    } else if (e.key === 'Tab' && e.shiftKey) {
      e.preventDefault();
      onOutdent?.();
    }
  }
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex a11y_no_noninteractive_element_interactions -->
<div
  bind:this={el}
  tabindex="0"
  class="divider-block my-3 px-1 outline-none rounded hover:bg-base-200/50 focus:bg-base-200/50 transition-colors cursor-default"
  onkeydown={handleKeydown}
  aria-label="Divider"
>
  <hr class="border-t border-base-300" />
</div>
