<script lang="ts">
  import { authStore } from '$lib/stores/auth.svelte';
  import { workspaceStore } from '$lib/stores/workspaces.svelte';
  import { goto } from '$app/navigation';

  let email = $state('');
  let password = $state('');
  let name = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    error = '';
    loading = true;
    try {
      await authStore.signup(email, password, name);
      workspaceStore.load();
      await goto('/');
    } catch (err: any) {
      error = err.message ?? 'Signup failed';
    } finally {
      loading = false;
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200">
  <div class="card w-full max-w-sm bg-base-100 shadow-xl">
    <div class="card-body">
      <h2 class="card-title text-2xl mb-2">Sign up</h2>
      <form onsubmit={handleSubmit}>
        <input bind:value={name} type="text" placeholder="Name" class="input input-bordered w-full mb-3" required />
        <input bind:value={email} type="email" placeholder="Email" class="input input-bordered w-full mb-3" required />
        <input bind:value={password} type="password" placeholder="Password" class="input input-bordered w-full mb-3" required />
        {#if error}
          <div class="alert alert-error text-sm py-2 mb-3">{error}</div>
        {/if}
        <button type="submit" class="btn btn-primary w-full" disabled={loading}>
          {loading ? 'Creating account...' : 'Sign up'}
        </button>
      </form>
      <p class="text-sm text-center mt-4 text-base-content/60">
        Already have an account? <a href="/login" class="link link-primary">Log in</a>
      </p>
    </div>
  </div>
</div>
