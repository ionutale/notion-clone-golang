<script lang="ts">
  import '../app.css';
  import { authStore } from '$lib/stores/auth.svelte';
  import { workspaceStore } from '$lib/stores/workspaces.svelte';
  import { theme } from '$lib/stores/theme.svelte';
  import { onMount } from 'svelte';
  import { page } from '$app/state';
  import { goto } from '$app/navigation';

  let { children } = $props();

  const publicPaths = ['/login', '/signup'];

  onMount(async () => {
    await authStore.check();
    if (authStore.user) {
      await workspaceStore.load();
    }
    if (!authStore.user && !publicPaths.includes(page.url.pathname)) {
      await goto('/login');
    }
  });
</script>

<svelte:head>
  <title>Notion Clone</title>
  <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>📝</text></svg>">
</svelte:head>

<div data-theme={theme.effective} class="min-h-screen bg-base-200">
  <nav class="navbar bg-base-100 border-b border-base-300 px-4">
    <div class="flex-1">
      <span class="text-xl font-bold">Notion Clone</span>
    </div>
  </nav>

  {#if authStore.loading}
    <div class="flex justify-center items-center min-h-screen">
      <span class="loading loading-spinner loading-lg text-primary"></span>
    </div>
  {:else if !authStore.user && !publicPaths.includes(page.url.pathname)}
    <!-- will redirect via onMount -->
  {:else}
    <main class="p-6">
      {@render children()}
    </main>
  {/if}
</div>
