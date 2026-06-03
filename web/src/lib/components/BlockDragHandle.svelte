<script lang="ts">
  let {
    blockId,
    onDragStart,
    visible = false,
    onTouchStart,
    onTouchMove,
    onTouchEnd,
  }: {
    blockId: string;
    onDragStart: (e: DragEvent) => void;
    visible: boolean;
    onTouchStart?: (e: TouchEvent) => void;
    onTouchMove?: (e: TouchEvent) => void;
    onTouchEnd?: (e: TouchEvent) => void;
  } = $props();

  let longPressTimer: ReturnType<typeof setTimeout> | undefined;
  let draggingViaTouch = $state(false);

  function handleDragStart(e: DragEvent) {
    e.dataTransfer?.setData('text/plain', blockId);
    e.dataTransfer!.effectAllowed = 'move';
    onDragStart(e);
  }

  function handleTouchStart(e: TouchEvent) {
    longPressTimer = setTimeout(() => {
      draggingViaTouch = true;
      onTouchStart?.(e);
    }, 500);
  }

  function handleTouchMove(e: TouchEvent) {
    if (draggingViaTouch) {
      e.preventDefault();
      onTouchMove?.(e);
    } else {
      clearTimeout(longPressTimer);
    }
  }

  function handleTouchEnd(e: TouchEvent) {
    clearTimeout(longPressTimer);
    if (draggingViaTouch) {
      e.preventDefault();
      draggingViaTouch = false;
      onTouchEnd?.(e);
    }
  }
</script>

{#if visible}
  <span
    draggable="true"
    ondragstart={handleDragStart}
    ontouchstart={handleTouchStart}
    ontouchmove={handleTouchMove}
    ontouchend={handleTouchEnd}
    class="drag-handle cursor-grab active:cursor-grabbing text-base-content/20 hover:text-base-content/40 transition-colors px-0.5 select-none inline-flex"
    class:opacity-50={draggingViaTouch}
    role="button"
    tabindex="-1"
    aria-label="Drag to reorder"
  >
    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
      <circle cx="9" cy="5" r="1.5" />
      <circle cx="15" cy="5" r="1.5" />
      <circle cx="9" cy="12" r="1.5" />
      <circle cx="15" cy="12" r="1.5" />
      <circle cx="9" cy="19" r="1.5" />
      <circle cx="15" cy="19" r="1.5" />
    </svg>
  </span>
{/if}
