import { api } from '$lib/api';

interface Workspace {
  id: string;
  name: string;
  owner_id: string;
  created_at: string;
}

class WorkspaceStore {
  workspaces = $state<Workspace[]>([]);
  activeWorkspaceId = $state<string | null>(null);

  get activeWorkspace() {
    return this.workspaces.find(w => w.id === this.activeWorkspaceId) ?? null;
  }

  async load() {
    const ws: Workspace[] = await api.request('GET', '/workspaces');
    this.workspaces = ws;
    if (ws.length > 0 && !this.activeWorkspaceId) {
      this.activeWorkspaceId = ws[0].id;
    }
  }

  async create(name: string) {
    const ws: Workspace = await api.request('POST', '/workspaces', { name });
    this.workspaces = [...this.workspaces, ws];
    this.activeWorkspaceId = ws.id;
  }

  async switchWorkspace(id: string) {
    this.activeWorkspaceId = id;
  }
}

export const workspaceStore = new WorkspaceStore();
