<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { workspaceStore } from '$lib/stores/workspaces.svelte';
  import { authStore } from '$lib/stores/auth.svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import type { PageSummary } from '$lib/types';

  let pages = $state<PageSummary[]>([]);
  let loading = $state(true);
  let dropdownOpen = $state(false);
  let search = $state('');
  let activeId = $derived($page.params.id);

  let filtered = $derived(
    search ? pages.filter(p => p.title.toLowerCase().includes(search.toLowerCase())) : pages
  );

  async function loadPages() {
    loading = true;
    try {
      pages = await blockStore.listPages();
    } catch (e) {
      console.error('Failed to load pages', e);
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadPages();
  });

  async function createPage() {
    const page = await blockStore.createPage();
    pages = [...pages, { id: page.id, title: page.content?.title ?? 'Untitled', created_at: page.created_at }];
    goto(`/pages/${page.id}`);
  }

  async function deletePage(e: Event, id: string) {
    e.stopPropagation();
    if (!confirm('Delete this page?')) return;
    try {
      await blockStore.deleteBlock(id);
      pages = pages.filter(p => p.id !== id);
      if (activeId === id) goto('/');
    } catch (err) {
      console.error('Failed to delete page', err);
    }
  }

  let editingId = $state<string | null>(null);
  let editTitle = $state('');

  function startRename(id: string, title: string) {
    editingId = id;
    editTitle = title;
  }

  async function commitRename(id: string) {
    if (editTitle.trim()) {
      await blockStore.updateBlock(id, { content: { title: editTitle.trim() } });
      pages = pages.map(p => p.id === id ? { ...p, title: editTitle.trim() } : p);
    }
    editingId = null;
  }
</script>

<aside class="w-64 h-full bg-base-200 border-r border-base-300 flex flex-col">
  <div class="p-3 border-b border-base-300">
    <div class="relative mb-2">
      <button
        onclick={() => dropdownOpen = !dropdownOpen}
        class="w-full flex items-center justify-between px-3 py-2 bg-base-300 rounded-lg hover:bg-base-200 transition-colors text-sm font-medium"
      >
        <span>{workspaceStore.activeWorkspace?.name ?? 'Select workspace'}</span>
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {#if dropdownOpen}
        <div class="absolute top-full left-0 right-0 mt-1 bg-base-100 border border-base-300 rounded-lg shadow-xl z-50 py-1">
          {#each workspaceStore.workspaces as ws}
            <button
              onclick={() => { workspaceStore.switchWorkspace(ws.id); dropdownOpen = false; }}
              class:bg-base-200={ws.id === workspaceStore.activeWorkspaceId}
              class="w-full text-left px-3 py-2 text-sm hover:bg-base-200"
            >
              {ws.name}
            </button>
          {/each}
          <hr class="border-base-200 my-1">
          <button
            onclick={async () => { const name = prompt('Workspace name'); if (name) await workspaceStore.create(name); dropdownOpen = false; }}
            class="w-full text-left px-3 py-2 text-sm text-primary hover:bg-base-200"
          >
            + New workspace
          </button>
        </div>
      {/if}
    </div>
    <button onclick={createPage} class="btn btn-primary btn-sm w-full gap-2">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
      New Page
    </button>
  </div>

  <div class="p-2">
    <input
      type="text"
      placeholder="Search pages..."
      bind:value={search}
      class="input input-ghost input-xs w-full"
    />
  </div>

  <div class="flex-1 overflow-y-auto">
    {#if loading}
      <div class="flex justify-center p-4">
        <span class="loading loading-spinner loading-sm"></span>
      </div>
    {:else if filtered.length === 0}
      <div class="text-center text-sm text-base-content/40 p-4">
        {search ? 'No pages found' : 'No pages yet'}
      </div>
    {:else}
      <ul class="menu menu-sm p-1">
        {#each filtered as p (p.id)}
          <li>
            <div
              class="flex items-center gap-2 rounded-lg"
              class:active={p.id === activeId}
            >
              {#if editingId === p.id}
                <input
                  type="text"
                  bind:value={editTitle}
                  onblur={() => commitRename(p.id)}
                  onkeydown={(e) => {
                    if (e.key === 'Enter') commitRename(p.id);
                    if (e.key === 'Escape') editingId = null;
                  }}
                  class="input input-xs input-ghost flex-1 min-w-0"
                  onclick={(e) => e.stopPropagation()}
                />
              {:else}
                <a
                  href="/pages/{p.id}"
                  class="flex-1 truncate"
                  ondblclick={() => startRename(p.id, p.title)}
                >
                  <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  {p.title}
                </a>
              {/if}
              <button
                onclick={(e) => deletePage(e, p.id)}
                class="btn btn-ghost btn-xs px-1 opacity-0 group-hover:opacity-100 hover:opacity-100"
                title="Delete page"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <div class="p-3 border-t border-base-300">
    <button
      onclick={async () => { await authStore.logout(); goto('/login'); }}
      class="w-full text-left px-3 py-2 text-sm text-base-content/50 hover:text-error transition-colors"
    >
      Log out
    </button>
  </div>
</aside>
