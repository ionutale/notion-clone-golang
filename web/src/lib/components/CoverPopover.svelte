<script lang="ts">
  import { api } from '$lib/api';

  const COLORS = ['#e5e7eb', '#fef3c7', '#dbeafe', '#d1fae5', '#fce7f3', '#ede9fe', '#ffe4e6', '#f0fdf4'];

  let {
    onselect = (_: { type: 'image' | 'color'; value: string }) => {},
    onremove = () => {},
    onclose = () => {},
  }: {
    onselect?: (detail: { type: 'image' | 'color'; value: string }) => void;
    onremove?: () => void;
    onclose?: () => void;
  } = $props();

  let tab = $state<'image' | 'color'>('image');
  let imageUrl = $state('');
  let uploading = $state(false);

  async function handleFileUpload(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    uploading = true;
    try {
      const { url } = await api.uploadFile(file);
      onselect({ type: 'image', value: url });
      onclose();
    } catch {
      // silent
    } finally {
      uploading = false;
      input.value = '';
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') onclose?.();
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div
  class="absolute z-50 mt-1 w-72 bg-base-100 border border-base-300 rounded-xl shadow-xl"
  onclick={(e) => e.stopPropagation()}
>
  <div class="flex border-b border-base-300" role="tablist">
    <button role="tab" aria-selected={tab === 'image'} onclick={() => tab = 'image'}
      class="flex-1 px-3 py-2 text-sm font-medium transition-colors"
      class:border-b-2={tab === 'image'} class:border-primary={tab === 'image'}
      class:text-primary={tab === 'image'}>Image</button>
    <button role="tab" aria-selected={tab === 'color'} onclick={() => tab = 'color'}
      class="flex-1 px-3 py-2 text-sm font-medium transition-colors"
      class:border-b-2={tab === 'color'} class:border-primary={tab === 'color'}
      class:text-primary={tab === 'color'}>Color</button>
  </div>

  <div role="tabpanel" class="p-3">
    {#if tab === 'image'}
      <div class="space-y-3">
        <label for="cover-upload-input" class="btn btn-outline btn-sm w-full" class:btn-disabled={uploading}>
          {uploading ? 'Uploading...' : 'Upload image'}
        </label>
        <input id="cover-upload-input" type="file" accept="image/*" onchange={handleFileUpload} class="hidden" disabled={uploading} />
        <div class="divider text-xs text-base-content/40">or paste URL</div>
        <div class="flex gap-2">
          <input bind:value={imageUrl} type="url" placeholder="https://example.com/banner.jpg"
            class="input input-bordered input-sm flex-1" />
          <button onclick={() => { if (imageUrl.trim()) { onselect({ type: 'image', value: imageUrl.trim() }); onclose(); }}}
            class="btn btn-primary btn-sm">Set</button>
        </div>
      </div>
    {:else}
      <div class="grid grid-cols-4 gap-2">
        {#each COLORS as color}
          <button
            onclick={() => { onselect({ type: 'color', value: color }); onclose(); }}
            class="w-full aspect-video rounded-lg border border-base-300 hover:scale-105 transition-transform"
            style="background-color: {color}"
          ></button>
        {/each}
      </div>
    {/if}
  </div>

  <div class="border-t border-base-300 p-1 flex justify-between">
    <button onclick={() => { onremove(); onclose(); }} class="btn btn-ghost btn-xs text-base-content/40 hover:text-error">Remove</button>
    <button onclick={onclose} class="btn btn-ghost btn-xs">Close</button>
  </div>
</div>
