<script lang="ts">
  let {
    open = false,
    title = 'Enter value',
    placeholder = '',
    confirmText = 'OK',
    cancelText = 'Cancel',
    onConfirm = ((_value: string) => {}) as (value: string) => void,
    onCancel = (() => {}) as () => void,
  } = $props();

  let value = $state('');
</script>

{#if open}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    onclick={onCancel}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
  >
    <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
    <div
      class="bg-base-100 rounded-xl shadow-2xl p-6 max-w-sm w-full mx-4"
      onclick={(e) => e.stopPropagation()}
    >
      <h3 class="text-lg font-semibold mb-4">{title}</h3>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        bind:value
        type="text"
        {placeholder}
        class="input input-bordered w-full mb-6"
        autofocus
      />
      <div class="flex justify-end gap-2">
        <button onclick={(e) => { e.stopPropagation(); e.preventDefault(); onCancel(); }} class="btn btn-ghost btn-sm">{cancelText}</button>
        <button onclick={() => onConfirm(value)} class="btn btn-primary btn-sm">{confirmText}</button>
      </div>
    </div>
  </div>
{/if}
