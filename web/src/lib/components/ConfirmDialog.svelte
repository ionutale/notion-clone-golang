<script lang="ts">
  let {
    open = false,
    title = 'Confirm',
    message = '',
    confirmText = 'OK',
    cancelText = 'Cancel',
    onConfirm = (() => {}) as () => void,
    onCancel = (() => {}) as () => void,
    variant = 'primary' as 'primary' | 'danger',
  } = $props();
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
      <h3 class="text-lg font-semibold mb-2">{title}</h3>
      <p class="text-sm text-base-content/70 mb-6">{message}</p>
      <div class="flex justify-end gap-2">
        <button onclick={onCancel} class="btn btn-ghost btn-sm">{cancelText}</button>
        <button
          onclick={onConfirm}
          class={['btn btn-sm', { 'btn-primary': variant === 'primary', 'btn-error': variant === 'danger' }]}
        >
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}
