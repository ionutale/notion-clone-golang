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

  let el = $state<HTMLButtonElement>();

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      e.stopPropagation();
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

<button
  bind:this={el}
  class="divider-block my-3 px-1 outline-none rounded hover:bg-base-200/50 focus:bg-base-200/50 transition-colors cursor-default w-full text-left"
  onkeydown={handleKeydown}
  aria-label="Divider"
>
  <hr class="border-t border-base-300 pointer-events-none" />
</button>
