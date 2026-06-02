<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { goto } from '$app/navigation';
  import Sidebar from '$lib/components/Sidebar.svelte';

  let health = $state<string>('checking...');
  let creating = $state(false);

  async function checkHealth() {
    try {
      const res = await fetch('/api/v1/health');
      const data = await res.json();
      health = data.status;
    } catch {
      health = 'offline';
    }
  }

  $effect(() => { checkHealth(); });

  async function createFirstPage() {
    creating = true;
    try {
      const page = await blockStore.createPage();
      goto(`/pages/${page.id}`);
    } catch (e) {
      creating = false;
    }
  }
</script>

<div class="flex h-[calc(100vh-4rem)]">
  <Sidebar />

  <div class="flex-1 flex items-center justify-center">
    <div class="max-w-md text-center">
      <h1 class="text-5xl font-bold mb-4 text-base-content">Notion Clone</h1>
      <p class="text-lg text-base-content/70 mb-8">
        A block-based document editor built with Go + SvelteKit
      </p>

      <div class="badge badge-lg gap-2 mb-8">
        <span class="w-2 h-2 rounded-full" class:bg-success={health !== 'offline'} class:bg-error={health === 'offline'}></span>
        API: {health}
      </div>

      <div class="space-y-4">
        <button
          onclick={createFirstPage}
          class="btn btn-primary btn-lg w-full"
          disabled={creating}
        >
          {#if creating}
            <span class="loading loading-spinner"></span>
          {/if}
          Create your first page
        </button>
      </div>

      <div class="mt-8 flex gap-2 justify-center flex-wrap">
        <span class="badge badge-outline">Go 1.26</span>
        <span class="badge badge-outline">Svelte 5</span>
        <span class="badge badge-outline">PostgreSQL 17</span>
        <span class="badge badge-outline">DaisyUI</span>
      </div>
    </div>
  </div>
</div>
