<script lang="ts">
  import { api } from '$lib/api';
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import type { SearchResult } from '$lib/types';

  let query = $state(page.url.searchParams.get('q') || '');
  let results = $state.raw<SearchResult[]>([]);
  let loading = $state(false);
  let searched = $state(false);
  let error = $state<string | null>(null);

  let debounceTimer: ReturnType<typeof setTimeout>;

  async function doSearch(q: string) {
    if (!q.trim()) {
      results = [];
      searched = false;
      return;
    }

    await goto(`/search?q=${encodeURIComponent(q)}`, { replaceState: true, keepFocus: true });

    loading = true;
    error = null;
    api.search(q)
      .then(res => {
        results = res;
        searched = true;
      })
      .catch(e => {
        error = e.message ?? 'Search failed';
        results = [];
      })
      .finally(() => {
        loading = false;
      });
  }

  function handleInput(e: Event) {
    const q = (e.target as HTMLInputElement).value;
    query = q;
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => doSearch(q), 300);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      clearTimeout(debounceTimer);
      doSearch(query);
    }
  }

  $effect(() => {
    const q = page.url.searchParams.get('q') || '';
    if (q) {
      query = q;
      doSearch(q);
    }
  });

  function blockTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      text: 'Text', heading_1: 'Heading 1', heading_2: 'Heading 2', heading_3: 'Heading 3',
      bullet_list_item: 'Bullet list', numbered_list_item: 'Numbered list',
      toggle: 'Toggle', divider: 'Divider', image: 'Image', page: 'Page',
    };
    return labels[type] ?? type;
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="max-w-3xl mx-auto py-8 px-4">
  <div class="mb-6">
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="search"
      placeholder="Search across all pages..."
      value={query}
      oninput={handleInput}
      class="input input-bordered input-lg w-full text-lg"
      autofocus
    />
  </div>

  {#if loading}
    <div class="flex justify-center py-20">
      <span class="loading loading-spinner loading-lg text-primary"></span>
    </div>
  {:else if error}
    <div class="alert alert-error shadow-lg my-8">
      <span>{error}</span>
    </div>
  {:else if searched && results.length === 0}
    <div class="text-center py-20 text-base-content/40">
      <p class="text-lg">No results for "{query}"</p>
      <p class="text-sm mt-1">Try different keywords</p>
    </div>
  {:else if searched}
    <div class="space-y-0.5">
      {#each results as r (r.block_id)}
        <a
          href="/pages/{r.page_id}"
          class="block px-4 py-3 rounded-lg hover:bg-base-200 transition-colors"
        >
          <div class="flex items-center gap-2 mb-1">
            <span class="text-sm font-medium">{r.page_title}</span>
            <span class="text-xs text-base-content/40 bg-base-300 px-1.5 py-0.5 rounded">{blockTypeLabel(r.block_type)}</span>
          </div>
          <p class="text-sm text-base-content/60 line-clamp-2">{r.excerpt}</p>
        </a>
      {/each}
    </div>
    <p class="text-xs text-base-content/40 text-center mt-4">{results.length} result{results.length !== 1 ? 's' : ''}</p>
  {:else}
    <div class="text-center py-20 text-base-content/40">
      <p class="text-lg">Search across all pages</p>
      <p class="text-sm mt-1">Type above to find content across your workspace</p>
    </div>
  {/if}
</div>
