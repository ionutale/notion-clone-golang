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
    const result = await api.uploadFile(file);
    onselect?.({ type: 'image', value: result.url });
    onclose?.();
  }

  function setImageUrl() {
    if (!urlInput.trim()) return;
    onselect?.({ type: 'image', value: urlInput.trim() });
    onclose?.();
  }
</script>

<div
  class="absolute z-50 w-72 rounded-xl shadow-xl border border-base-300 bg-base-100 overflow-hidden"
  onclick={(e) => e.stopPropagation()}
>
  <div class="flex border-b border-base-200">
    <button
      class="flex-1 px-4 py-2.5 text-sm font-medium transition-colors"
      class:text-primary={tab === 'emoji'}
      class:border-b-2={tab === 'emoji'}
      class:border-primary={tab === 'emoji'}
      onclick={() => tab = 'emoji'}
    >
      Emoji
    </button>
    <button
      class="flex-1 px-4 py-2.5 text-sm font-medium transition-colors"
      class:text-primary={tab === 'image'}
      class:border-b-2={tab === 'image'}
      class:border-primary={tab === 'image'}
      onclick={() => tab = 'image'}
    >
      Image
    </button>
  </div>

  <div class="p-3">
    {#if tab === 'emoji'}
      <div class="grid grid-cols-5 gap-1">
        {#each emojis as emoji}
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
        <label class="flex">
          <input
            type="file"
            accept="image/*"
            class="hidden"
            onchange={handleUpload}
            id="icon-upload-input"
          />
          <button
            class="btn btn-outline btn-sm w-full"
            onclick={() => document.getElementById('icon-upload-input')?.click()}
          >
            Upload Image
          </button>
        </label>

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
