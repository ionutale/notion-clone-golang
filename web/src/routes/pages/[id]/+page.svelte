<script lang="ts">
  import { page } from '$app/state';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import Editor from '$lib/components/Editor.svelte';

  let pageId = $derived(page.params.id!);
</script>

<div class="flex h-[calc(100vh-4rem)]">
  <Sidebar />

  <div class="flex-1 overflow-y-auto">
    <Editor {pageId} />
    <!-- <div>ciao</div> -->
  </div>
</div>
