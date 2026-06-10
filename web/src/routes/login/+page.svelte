<script lang="ts">
  import { authStore } from '$lib/stores/auth.svelte';
  import { workspaceStore } from '$lib/stores/workspaces.svelte';
  import { goto } from '$app/navigation';

  let email = $state('');
  let password = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      await authStore.login(email, password);
      workspaceStore.load();
      await goto('/');
    } catch (err: any) {
      error = err.message ?? 'Login failed';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200">
  <div class="card w-full max-w-sm bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title text-2xl mb-2">Log in</h2>
      <form onsubmit={handleSubmit}>
        <input bind:value={email} type="email" placeholder="Email" class="input input-bordered w-full mb-3" required />
        <input bind:value={password} type="password" placeholder="Password" class="input input-bordered w-full mb-3" required />
        {#if error}
          <div class="alert alert-error text-sm py-2 mb-3">{error}</div>
        {/if}
        <button type="submit" class="btn btn-primary w-full" disabled={loading}>
          {loading ? 'Logging in...' : 'Log in'}
        </button>
      </form>
      <p class="text-sm text-center mt-4 text-base-content/60">
        Don't have an account? <a href="/signup" class="link link-primary">Sign up</a>
      </p>
    </div>
  </div>
</div>
