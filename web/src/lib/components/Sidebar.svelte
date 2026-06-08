<script lang="ts">
  import { blockStore } from '$lib/stores/blocks.svelte';
  import { workspaceStore } from '$lib/stores/workspaces.svelte';
  import { authStore } from '$lib/stores/auth.svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { api } from '$lib/api';
  import type { PageSummary } from '$lib/types';

  interface Member {
    user_id: string;
    email: string;
    name: string | null;
    role: string;
    joined_at: string;
  }

  let inviteEmail = $state('');
  let inviting = $state(false);
  let workspaceMembers = $state<Member[]>([]);

  async function loadMembers() {
    if (!workspaceStore.activeWorkspaceId) return;
    try {
      workspaceMembers = await api.request('GET', `/workspaces/${workspaceStore.activeWorkspaceId}/members`);
    } catch {
      workspaceMembers = [];
    }
  }

  async function handleInvite() {
    if (!inviteEmail.trim() || !workspaceStore.activeWorkspaceId) return;
    inviting = true;
    try {
      await api.request('POST', `/workspaces/${workspaceStore.activeWorkspaceId}/members`, {
        email: inviteEmail.trim(),
        role: 'member',
      });
      inviteEmail = '';
      await loadMembers();
    } catch (err: any) {
      alert(err.message ?? 'Failed to invite');
    } finally {
      inviting = false;
    }
  }

  let pages = $state<PageSummary[]>([]);
  let loading = $state(true);
  let dropdownOpen = $state(false);
  let search = $state('');
  let activeId = $derived($page.params.id);

  let filtered = $derived(
    search ? pages.filter(p => p.title.toLowerCase().includes(search.toLowerCase())) : pages
  );

  let draggedId = $state<string | null>(null);
  let dragOverId = $state<string | null>(null);
  let dragPosition = $state<'before' | 'after' | null>(null);

  let touchDragging = $state(false);
  let touchTimer: ReturnType<typeof setTimeout> | undefined;
  let touchStartY = $state(0);
  let touchStartX = $state(0);
  let draggedEl: HTMLElement | null = null;

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
    blockStore.loadFavorites();
  });

  $effect(() => {
    if (workspaceStore.activeWorkspaceId) {
      loadMembers();
    }
  });

  let favoritePages = $derived(pages.filter(p => blockStore.favoriteIds.has(p.id)));

  async function createPage() {
    const page = await blockStore.createPage();
    pages = [...pages, { id: page.id, title: page.content?.title ?? 'Untitled', icon: page.content?.icon, icon_type: page.content?.icon_type, position: page.position, created_at: page.created_at }];
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

  function calcDropPosition(
    pages: PageSummary[], draggedId: string, dropId: string, position: 'before' | 'after'
  ): { newPos: number; fromIdx: number; toIdx: number } {
    const fromIdx = pages.findIndex(p => p.id === draggedId);
    const toIdx = pages.findIndex(p => p.id === dropId);
    let newPos: number;
    if (position === 'before') {
      if (toIdx === 0) {
        newPos = pages[0].position / 2;
      } else {
        newPos = (pages[toIdx - 1].position + pages[toIdx].position) / 2;
      }
    } else {
      if (toIdx === pages.length - 1) {
        newPos = pages[pages.length - 1].position + 1;
      } else if (toIdx + 1 === fromIdx) {
        newPos = (pages[toIdx].position + pages[toIdx + 2].position) / 2;
      } else {
        newPos = (pages[toIdx].position + pages[toIdx + 1].position) / 2;
      }
    }
    return { newPos, fromIdx, toIdx };
  }

  function optimisticReorder(
    pages: PageSummary[], draggedId: string, dropId: string, newPos: number, position: 'before' | 'after'
  ): PageSummary[] {
    const fromIdx = pages.findIndex(p => p.id === draggedId);
    const toIdx = pages.findIndex(p => p.id === dropId);
    const item = pages[fromIdx];
    const updated = pages.filter(p => p.id !== draggedId);
    const adjustedToIdx = fromIdx < toIdx ? toIdx - 1 : toIdx;
    const insertAt = position === 'before' ? adjustedToIdx : adjustedToIdx + 1;
    updated.splice(insertAt, 0, { ...item, position: newPos });
    return updated;
  }

  function handleDragStart(e: DragEvent, id: string) {
    if (search) return;
    draggedId = id;
    e.dataTransfer!.setData('text/plain', id);
    e.dataTransfer!.effectAllowed = 'move';
    requestAnimationFrame(() => {
      (e.currentTarget as HTMLElement).classList.add('opacity-50');
    });
  }

  function handleDragOver(e: DragEvent, id: string) {
    if (!draggedId || draggedId === id || search) return;
    e.preventDefault();
    e.dataTransfer!.dropEffect = 'move';
    const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
    const mid = rect.top + rect.height / 2;
    dragOverId = id;
    dragPosition = e.clientY < mid ? 'before' : 'after';
  }

  function handleDrop(e: DragEvent, dropId: string) {
    e.preventDefault();
    if (!draggedId || draggedId === dropId || search || !dragPosition) {
      dragOverId = null;
      dragPosition = null;
      draggedId = null;
      return;
    }
    const { newPos } = calcDropPosition(pages, draggedId, dropId, dragPosition);
    pages = optimisticReorder(pages, draggedId, dropId, newPos, dragPosition);
    blockStore.moveBlock(draggedId, null, newPos).catch(() => loadPages());
    dragOverId = null;
    dragPosition = null;
    draggedId = null;
  }

  function handleDragEnd() {
    dragOverId = null;
    dragPosition = null;
    draggedId = null;
  }

  function handleTouchStart(e: TouchEvent, id: string) {
    if (search || editingId === id) return;
    touchStartY = e.touches[0].clientY;
    touchStartX = e.touches[0].clientX;
    draggedEl = e.currentTarget as HTMLElement;
    touchTimer = setTimeout(() => {
      touchDragging = true;
      draggedId = id;
      if (draggedEl) draggedEl.classList.add('opacity-50');
    }, 500);
  }

  function handleTouchMove(e: TouchEvent) {
    if (!touchDragging) {
      const dy = Math.abs(e.touches[0].clientY - touchStartY);
      const dx = Math.abs(e.touches[0].clientX - touchStartX);
      if (dy > 10 || dx > 10) clearTimeout(touchTimer);
      return;
    }
    e.preventDefault();
    const touchY = e.touches[0].clientY;
    const items = document.querySelectorAll<HTMLElement>('[data-page-id]');
    let found = false;
    for (const item of items) {
      const rect = item.getBoundingClientRect();
      if (touchY >= rect.top && touchY <= rect.bottom) {
        const id = item.dataset.pageId!;
        if (id !== draggedId) {
          dragOverId = id;
          dragPosition = touchY < rect.top + rect.height / 2 ? 'before' : 'after';
        }
        found = true;
        break;
      }
    }
    if (!found) {
      dragOverId = null;
      dragPosition = null;
    }
  }

  function handleTouchEnd() {
    clearTimeout(touchTimer);
    if (!touchDragging) return;
    touchDragging = false;
    if (draggedId && dragOverId && dragPosition) {
      const { newPos } = calcDropPosition(pages, draggedId, dragOverId, dragPosition);
      pages = optimisticReorder(pages, draggedId, dragOverId, newPos, dragPosition);
      blockStore.moveBlock(draggedId, null, newPos).catch(() => loadPages());
    }
    if (draggedEl) draggedEl.classList.remove('opacity-50');
    dragOverId = null;
    dragPosition = null;
    draggedId = null;
    draggedEl = null;
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
          <hr class="border-base-200 my-1">
          <div class="px-3 py-2">
            <p class="text-xs font-medium text-base-content/40 uppercase mb-2">Members</p>
            {#if workspaceMembers.length > 0}
              {#each workspaceMembers as m}
                <div class="flex items-center justify-between py-1">
                  <span class="text-sm truncate">{m.name || m.email}</span>
                  <span class="text-xs text-base-content/40 shrink-0">{m.role}</span>
                </div>
              {/each}
            {:else}
              <p class="text-xs text-base-content/30">No members yet</p>
            {/if}
            <div class="flex gap-2 mt-2">
              <input
                bind:value={inviteEmail}
                type="email"
                placeholder="Email to invite"
                class="input input-ghost input-xs flex-1 min-w-0"
              />
              <button
                onclick={handleInvite}
                class="btn btn-primary btn-xs"
                disabled={!inviteEmail.trim() || inviting}
              >
                {inviting ? '...' : 'Invite'}
              </button>
            </div>
          </div>
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

  {#if favoritePages.length > 0}
    <div class="px-2 pt-2 pb-1">
      <p class="text-xs font-medium text-base-content/40 uppercase tracking-wider px-3">Favorites</p>
    </div>
    <ul class="menu menu-sm p-1">
      {#each favoritePages as p (p.id)}
        <li>
          <div class="flex items-center gap-2 rounded-lg" class:active={p.id === activeId}>
            <a href="/pages/{p.id}" class="flex-1 truncate flex items-center gap-1.5">
              {#if p.icon_type === 'image'}
                <img src={p.icon} alt="" class="w-4 h-4 rounded object-cover shrink-0" />
              {:else if p.icon}
                <span class="text-sm shrink-0">{p.icon}</span>
              {:else}
                <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              {/if}
              {p.title}
            </a>
            <button
              onclick={async () => { await blockStore.toggleFavorite(p.id); }}
              class="btn btn-ghost btn-xs px-1 text-warning"
              title="Unfavorite"
            >
              <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
              </svg>
            </button>
          </div>
        </li>
      {/each}
    </ul>
    <div class="divider my-1"></div>
  {/if}

  <div class="p-2">
    <input
      type="text"
      placeholder="Search pages..."
      bind:value={search}
      onkeydown={(e) => {
        if (e.key === 'Enter' && search.trim()) {
          goto(`/search?q=${encodeURIComponent(search.trim())}`);
        }
      }}
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
          <li
            draggable={search === '' && editingId !== p.id}
            ondragstart={(e) => handleDragStart(e, p.id)}
            ondragover={(e) => handleDragOver(e, p.id)}
            ondrop={(e) => handleDrop(e, p.id)}
            ondragend={handleDragEnd}
            ontouchstart={(e) => handleTouchStart(e, p.id)}
            ontouchmove={handleTouchMove}
            ontouchend={handleTouchEnd}
            data-page-id={p.id}
            class="group"
            class:opacity-50={draggedId === p.id}
          >
            {#if dragOverId === p.id && dragPosition === 'before'}
              <div class="h-0.5 bg-primary rounded-full mb-0.5"></div>
            {/if}
            <div
              class="flex items-center gap-2 rounded-lg"
              class:active={p.id === activeId}
            >
              {#if editingId !== p.id}
                <span class="drag-handle cursor-grab text-base-content/20 hover:text-base-content/40 transition-colors px-0.5 select-none opacity-0 group-hover:opacity-100">
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
                  class="flex-1 truncate flex items-center gap-1.5"
                  ondblclick={() => startRename(p.id, p.title)}
                >
                  {#if p.icon_type === 'image'}
                    <img src={p.icon} alt="" class="w-4 h-4 rounded object-cover shrink-0" />
                  {:else if p.icon}
                    <span class="text-sm shrink-0">{p.icon}</span>
                  {:else}
                    <svg class="w-4 h-4 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  {/if}
                  {p.title}
                </a>
              {/if}
              <button
                onclick={async () => { await blockStore.toggleFavorite(p.id); }}
                class="btn btn-ghost btn-xs px-1 opacity-0 group-hover:opacity-100 hover:opacity-100"
                class:text-warning={blockStore.favoriteIds.has(p.id)}
                title={blockStore.favoriteIds.has(p.id) ? 'Unfavorite' : 'Favorite'}
              >
                <svg class="w-3.5 h-3.5" fill={blockStore.favoriteIds.has(p.id) ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
                </svg>
              </button>
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
            {#if dragOverId === p.id && dragPosition === 'after'}
              <div class="h-0.5 bg-primary rounded-full mt-0.5"></div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <div class="p-3 border-t border-base-300 space-y-1">
    <a href="/settings" class="flex items-center gap-2 px-3 py-2 text-sm text-base-content/50 hover:text-base-content transition-colors rounded-lg hover:bg-base-200">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
      Settings
    </a>
    <a href="/trash" class="flex items-center gap-2 px-3 py-2 text-sm text-base-content/50 hover:text-base-content transition-colors rounded-lg hover:bg-base-200">
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
      </svg>
      Trash
    </a>
    <button
      onclick={async () => { await authStore.logout(); goto('/login'); }}
      class="w-full text-left px-3 py-2 text-sm text-base-content/50 hover:text-error transition-colors"
    >
      Log out
    </button>
  </div>
</aside>
