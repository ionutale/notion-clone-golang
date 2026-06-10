<script lang="ts">
  import { api } from '$lib/api';
  import { authStore } from '$lib/stores/auth.svelte';
  import { theme } from '$lib/stores/theme.svelte';
  import { goto } from '$app/navigation';
  import PromptDialog from '$lib/components/PromptDialog.svelte';

  let name = $state(authStore.user?.name ?? '');
  let email = $state(authStore.user?.email ?? '');
  let profileSaving = $state(false);
  let profileError = $state('');
  let profileSuccess = $state(false);

  let currentPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');
  let passwordSaving = $state(false);
  let passwordError = $state('');
  let passwordSuccess = $state(false);

  let showDeleteConfirm = $state(false);
  let deleteConfirmEmail = $state('');
  let deleteSaving = $state(false);
  let deleteError = $state('');

  let showEmailPasswordPrompt = $state(false);
  let showDeletePasswordPrompt = $state(false);
  let pendingProfilePassword = $state('');

  async function goBack() {
    await goto('/');
  }

  async function handleEmailChangeSubmit() {
    if (!pendingProfilePassword) return;
    profileSaving = true;
    profileError = '';
    profileSuccess = false;
    try {
      const user = await api.updateProfile({
        name,
        email,
        current_password: pendingProfilePassword,
      });
      authStore.user = user;
      profileSuccess = true;
    } catch (err: any) {
      profileError = err.message ?? 'Failed to update profile';
    } finally {
      profileSaving = false;
      pendingProfilePassword = '';
    }
  }

  async function saveProfile(e: Event) {
    e.preventDefault();
    if (email !== authStore.user?.email) {
      showEmailPasswordPrompt = true;
      return;
    }
    profileSaving = true;
    profileError = '';
    profileSuccess = false;
    try {
      const user = await api.updateProfile({ name, email });
      authStore.user = user;
      profileSuccess = true;
    } catch (err: any) {
      profileError = err.message ?? 'Failed to update profile';
    } finally {
      profileSaving = false;
    }
  }

  async function savePassword(e: Event) {
    e.preventDefault();
    passwordError = '';
    passwordSuccess = false;
    if (newPassword.length < 8) { passwordError = 'Password must be at least 8 characters'; return; }
    if (newPassword !== confirmPassword) { passwordError = 'Passwords do not match'; return; }
    passwordSaving = true;
    try {
      await api.updatePassword({ current_password: currentPassword, new_password: newPassword });
      passwordSuccess = true;
      currentPassword = '';
      newPassword = '';
      confirmPassword = '';
    } catch (err: any) {
      passwordError = err.message ?? 'Failed to update password';
    } finally {
      passwordSaving = false;
    }
  }

  async function handleDeleteAccount(password: string) {
    if (!password) return;
    if (deleteConfirmEmail !== authStore.user?.email) return;
    deleteSaving = true;
    deleteError = '';
    try {
      await api.deleteAccount({ password });
      authStore.user = null;
      authStore.accessToken = null;
      await goto('/login');
    } catch (err: any) {
      deleteError = err.message ?? 'Failed to delete account';
    } finally {
      deleteSaving = false;
    }
  }
</script>

<PromptDialog
  open={showEmailPasswordPrompt}
  title="Enter your current password"
  placeholder="Current password"
  confirmText="Confirm"
  onConfirm={(pw: string) => {
    showEmailPasswordPrompt = false;
    pendingProfilePassword = pw;
    handleEmailChangeSubmit();
  }}
  onCancel={() => showEmailPasswordPrompt = false}
/>

<PromptDialog
  open={showDeletePasswordPrompt}
  title="Enter your password to confirm deletion"
  placeholder="Password"
  confirmText="Delete"
  onConfirm={(pw: string) => {
    showDeletePasswordPrompt = false;
    handleDeleteAccount(pw);
  }}
  onCancel={() => showDeletePasswordPrompt = false}
/>

<div class="max-w-2xl mx-auto py-8 px-4">
  <div class="flex items-center gap-4 mb-8">
    <button onclick={goBack} class="btn btn-ghost btn-sm">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
      Back
    </button>
    <h1 class="text-2xl font-bold">Settings</h1>
  </div>

  <div class="card bg-base-100 border border-base-300 mb-6">
    <div class="card-body">
      <h2 class="card-title text-lg mb-4">Profile</h2>
      <form onsubmit={saveProfile}>
        <div class="form-control mb-3">
          <label class="label" for="settings-name"><span class="label-text">Name</span></label>
          <input id="settings-name" bind:value={name} type="text" class="input input-bordered" required />
        </div>
        <div class="form-control mb-3">
          <label class="label" for="settings-email"><span class="label-text">Email</span></label>
          <input id="settings-email" bind:value={email} type="email" class="input input-bordered" required />
        </div>
        {#if profileError}<div class="alert alert-error text-sm py-2 mb-3">{profileError}</div>{/if}
        {#if profileSuccess}<div class="alert alert-success text-sm py-2 mb-3">Profile updated</div>{/if}
        <button type="submit" class="btn btn-primary" disabled={profileSaving}>
          {profileSaving ? 'Saving...' : 'Save'}
        </button>
      </form>
    </div>
  </div>

  <div class="card bg-base-100 border border-base-300 mb-6">
    <div class="card-body">
      <h2 class="card-title text-lg mb-4">Password</h2>
      <form onsubmit={savePassword}>
        <div class="form-control mb-3">
          <label class="label" for="settings-current-pw"><span class="label-text">Current password</span></label>
          <input id="settings-current-pw" bind:value={currentPassword} type="password" class="input input-bordered" required />
        </div>
        <div class="form-control mb-3">
          <label class="label" for="settings-new-pw"><span class="label-text">New password</span></label>
          <input id="settings-new-pw" bind:value={newPassword} type="password" class="input input-bordered" required minlength={8} />
        </div>
        <div class="form-control mb-3">
          <label class="label" for="settings-confirm-pw"><span class="label-text">Confirm new password</span></label>
          <input id="settings-confirm-pw" bind:value={confirmPassword} type="password" class="input input-bordered" required minlength={8} />
        </div>
        {#if passwordError}<div class="alert alert-error text-sm py-2 mb-3">{passwordError}</div>{/if}
        {#if passwordSuccess}<div class="alert alert-success text-sm py-2 mb-3">Password updated</div>{/if}
        <button type="submit" class="btn btn-primary" disabled={passwordSaving}>
          {passwordSaving ? 'Changing...' : 'Change password'}
        </button>
      </form>
    </div>
  </div>

  <div class="card bg-base-100 border border-base-300 mb-6">
    <div class="card-body">
      <h2 class="card-title text-lg mb-4">Theme</h2>
      <div class="flex flex-col gap-2">
        {#each ['light', 'dark', 'system'] as option}
          <label class="flex items-center gap-3 cursor-pointer">
            <input
              type="radio"
              name="theme"
              class="radio radio-primary"
              value={option}
              checked={theme.preference === option}
              onchange={() => theme.preference = option}
            />
            <span class="capitalize">{option}</span>
          </label>
        {/each}
      </div>
    </div>
  </div>

  <div class="card bg-base-100 border border-error mb-6">
    <div class="card-body">
      <h2 class="card-title text-lg text-error mb-4">Danger Zone</h2>
      {#if !showDeleteConfirm}
        <p class="text-sm text-base-content/60 mb-4">Permanently delete your account and all associated data. This cannot be undone.</p>
        <button onclick={() => showDeleteConfirm = true} class="btn btn-outline btn-error">Delete account</button>
      {:else}
        <p class="text-sm mb-3">Type <strong>{authStore.user?.email}</strong> to confirm:</p>
        <input bind:value={deleteConfirmEmail} type="email" placeholder={authStore.user?.email} class="input input-bordered w-full mb-3" />
        {#if deleteError}<div class="alert alert-error text-sm py-2 mb-3">{deleteError}</div>{/if}
        <div class="flex gap-2">
          <button onclick={() => { showDeleteConfirm = false; deleteConfirmEmail = ''; }} class="btn btn-ghost">Cancel</button>
          <button
            onclick={() => showDeletePasswordPrompt = true}
            class="btn btn-error"
            disabled={deleteConfirmEmail !== authStore.user?.email || deleteSaving}
          >
            {deleteSaving ? 'Deleting...' : 'Delete my account'}
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>
