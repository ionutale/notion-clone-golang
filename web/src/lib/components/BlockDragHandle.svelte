<script lang="ts">
  let { blockId, onDragStart, visible = false }:
    {
      blockId: string;
      onDragStart: (e: DragEvent) => void;
      visible: boolean;
    } = $props();

  function handleDragStart(e: DragEvent) {
    e.dataTransfer?.setData('text/plain', blockId);
    e.dataTransfer!.effectAllowed = 'move';
    onDragStart(e);
  }
</script>

{#if visible}
  <span
    draggable="true"
    ondragstart={handleDragStart}
    class="drag-handle cursor-grab active:cursor-grabbing text-base-content/20 hover:text-base-content/40 transition-colors px-0.5 select-none inline-flex"
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
