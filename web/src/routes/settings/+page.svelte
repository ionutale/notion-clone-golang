<script lang="ts">
  import { api } from '$lib/api';
  import { authStore } from '$lib/stores/auth.svelte';
  import { goto } from '$app/navigation';

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

  async function saveProfile(e: Event) {
    e.preventDefault();
    profileSaving = true;
    profileError = '';
    profileSuccess = false;
    try {
      const user = await api.updateProfile({
        name,
        email,
        current_password: email !== authStore.user?.email ? prompt('Enter current password to change email:') ?? undefined : undefined,
      });
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

  async function handleDeleteAccount() {
    if (deleteConfirmEmail !== authStore.user?.email) return;
    deleteSaving = true;
    deleteError = '';
    try {
      const password = prompt('Enter your password to confirm deletion:');
      if (!password) { deleteSaving = false; return; }
      await api.deleteAccount({ password });
      authStore.user = null;
      authStore.accessToken = null;
      goto('/login');
    } catch (err: any) {
      deleteError = err.message ?? 'Failed to delete account';
    } finally {
      deleteSaving = false;
    }
  }
</script>

<div class="max-w-2xl mx-auto py-8 px-4">
  <h1 class="text-2xl font-bold mb-8">Settings</h1>

  <div class="card bg-base-100 border border-base-300 mb-6">
    <div class="card-body">
      <h2 class="card-title text-lg mb-4">Profile</h2>
      <form onsubmit={saveProfile}>
        <div class="form-control mb-3">
          <label class="label"><span class="label-text">Name</span></label>
          <input bind:value={name} type="text" class="input input-bordered" required />
        </div>
        <div class="form-control mb-3">
          <label class="label"><span class="label-text">Email</span></label>
          <input bind:value={email} type="email" class="input input-bordered" required />
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
          <label class="label"><span class="label-text">Current password</span></label>
          <input bind:value={currentPassword} type="password" class="input input-bordered" required />
        </div>
        <div class="form-control mb-3">
          <label class="label"><span class="label-text">New password</span></label>
          <input bind:value={newPassword} type="password" class="input input-bordered" required minlength={8} />
        </div>
        <div class="form-control mb-3">
          <label class="label"><span class="label-text">Confirm new password</span></label>
          <input bind:value={confirmPassword} type="password" class="input input-bordered" required minlength={8} />
        </div>
        {#if passwordError}<div class="alert alert-error text-sm py-2 mb-3">{passwordError}</div>{/if}
        {#if passwordSuccess}<div class="alert alert-success text-sm py-2 mb-3">Password updated</div>{/if}
        <button type="submit" class="btn btn-primary" disabled={passwordSaving}>
          {passwordSaving ? 'Changing...' : 'Change password'}
        </button>
      </form>
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
            onclick={handleDeleteAccount}
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
