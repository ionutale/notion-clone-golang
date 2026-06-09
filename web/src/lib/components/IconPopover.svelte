<script lang="ts">
  import { api } from '$lib/api';

  let {
    onselect,
    onremove,
    onclose,
  }: {
    onselect?: (detail: { type: 'emoji' | 'image'; value: string }) => void;
    onremove?: () => void;
    onclose?: () => void;
  } = $props();

  let tab = $state<'emoji' | 'image'>('emoji');
  let urlInput = $state('');
  let uploading = $state(false);

  const emojis = [
    '😀', '🎉', '🚀', '💡', '📝', '🔥', '⭐', '💻', '🎯', '📚',
    '🌟', '💪', '🎨', '🏆', '📌', '🎵', '🔧', '💰', '🎈', '🌍',
    '🌈', '❤️', '🧠', '🎁', '📷', '🛠️', '🚧', '⚡', '🎮', '🏠',
    '💎', '🎸', '🌿', '🥇', '🎪', '🎭', '📂', '🔗', '⏰', '💼',
    '🔒', '📎',
  ];

  function selectEmoji(emoji: string) {
    onselect?.({ type: 'emoji', value: emoji });
    onclose?.();
  }

  async function handleUpload(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    uploading = true;
    try {
      const { url } = await api.uploadFile(file);
      onselect?.({ type: 'image', value: url });
      onclose?.();
    } catch {
      // upload failed silently - the parent can handle via toast
    } finally {
      uploading = false;
      input.value = '';
    }
  }

  function setImageUrl() {
    if (!urlInput.trim()) return;
    onselect?.({ type: 'image', value: urlInput.trim() });
    onclose?.();
  }
  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose?.();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div
  class="absolute z-50 w-72 rounded-xl shadow-xl border border-base-300 bg-base-100 overflow-hidden"
  role="dialog"
  tabindex="-1"
  onclick={(e) => e.stopPropagation()}
  onkeydown={handleKeydown}
>
  <div class="flex border-b border-base-200">
    <button
      role="tab"
      aria-selected={tab === 'emoji'}
      aria-controls="icon-tabpanel"
      class={['flex-1 px-4 py-2.5 text-sm font-medium transition-colors', { 'text-primary': tab === 'emoji', 'border-b-2': tab === 'emoji', 'border-primary': tab === 'emoji' }]}
      onclick={() => tab = 'emoji'}
    >
      Emoji
    </button>
    <button
      role="tab"
      aria-selected={tab === 'image'}
      aria-controls="icon-tabpanel"
      class={['flex-1 px-4 py-2.5 text-sm font-medium transition-colors', { 'text-primary': tab === 'image', 'border-b-2': tab === 'image', 'border-primary': tab === 'image' }]}
      onclick={() => tab = 'image'}
    >
      Image
    </button>
  </div>

  <div role="tabpanel" id="icon-tabpanel" class="p-3">
    {#if tab === 'emoji'}
      <div class="grid grid-cols-5 gap-1">
        {#each emojis as emoji (emoji)}
          <button
            class="w-10 h-10 flex items-center justify-center text-lg rounded-lg hover:bg-base-200 transition-colors"
            onclick={() => selectEmoji(emoji)}
          >
            {emoji}
          </button>
        {/each}
      </div>
    {:else}
      <div class="space-y-3">
        <label for="icon-upload-input" class={['btn btn-outline btn-sm w-full', { 'btn-disabled': uploading }]}>
          {uploading ? 'Uploading...' : 'Upload image'}
        </label>
        <input id="icon-upload-input" type="file" accept="image/*" onchange={handleUpload} class="hidden" disabled={uploading} />

        <div class="flex items-center gap-2">
          <div class="flex-1 h-px bg-base-300"></div>
          <span class="text-xs text-base-content/40">or paste URL</span>
          <div class="flex-1 h-px bg-base-300"></div>
        </div>

        <div class="flex gap-2">
          <input
            type="text"
            placeholder="Paste image URL..."
            bind:value={urlInput}
            class="input input-bordered input-sm flex-1"
          />
          <button
            class="btn btn-primary btn-sm"
            onclick={setImageUrl}
            disabled={!urlInput.trim()}
          >
            Set
          </button>
        </div>
      </div>
    {/if}
  </div>

  <div class="flex items-center justify-between border-t border-base-200 px-3 py-2">
    <button
      class="btn btn-ghost btn-xs hover:text-error transition-colors"
      onclick={() => { onremove?.(); onclose?.(); }}
    >
      Remove
    </button>
    <button
      class="btn btn-ghost btn-xs"
      onclick={onclose}
    >
      Close
    </button>
  </div>
</div>
