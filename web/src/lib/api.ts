import { authStore } from '$lib/stores/auth.svelte';
import { workspaceStore } from '$lib/stores/workspaces.svelte';
import type { Block, BlockType, PageSummary, PageTree, SearchResult, User } from './types';

const BASE_URL = '/api/v1';

export class ApiError extends Error {
  constructor(public status: number, text: string) {
    super(text);
  }
}

class ApiClient {
  private _refreshing = false;
  private _wsLoadPromise: Promise<void> | null = null;

  private async ensureWorkspace(): Promise<void> {
    if (workspaceStore.activeWorkspaceId) return;
    if (workspaceStore.workspaces.length > 0) {
      workspaceStore.activeWorkspaceId = workspaceStore.workspaces[0].id;
      return;
    }
    if (!this._wsLoadPromise) {
      this._wsLoadPromise = workspaceStore.load();
    }
    await this._wsLoadPromise;
  }

  private async wsPrefix(): Promise<string> {
    await this.ensureWorkspace();
    return workspaceStore.activeWorkspaceId
      ? `/workspaces/${workspaceStore.activeWorkspaceId}`
      : '';
  }

  private needsWorkspace(path: string): boolean {
    return !path.startsWith('/auth') && !path.startsWith('/workspaces') && !path.startsWith('/uploads');
  }

  requestInner<T>(method: string, path: string, body?: any): Promise<T> {
    return this._requestInner<T>(method, path, body);
  }

  private async _requestInner<T>(method: string, path: string, body?: any): Promise<T> {
    if (this.needsWorkspace(path)) {
      await this.ensureWorkspace();
    }
    const opts: RequestInit = { method };
    if (body !== undefined) {
      opts.headers = { 'Content-Type': 'application/json' };
      opts.body = JSON.stringify(body);
    }
    if (authStore.accessToken) {
      opts.headers = { ...(opts.headers as Record<string, string>), 'Authorization': `Bearer ${authStore.accessToken}` };
    }
    const res = await fetch(`${BASE_URL}${path}`, opts);
    if (!res.ok) {
      const text = await res.text();
      let message = text;
      try { const json = JSON.parse(text); if (json.error) message = json.error; } catch { /* use raw text */ }
      throw new ApiError(res.status, message);
    }
    if (res.status === 204) return undefined as T;
    return res.json();
  }

  async request<T>(method: string, path: string, body?: any): Promise<T> {
    try {
      return await this._requestInner<T>(method, path, body);
    } catch (err: any) {
      if (err instanceof ApiError && err.status === 401 && !this._refreshing) {
        this._refreshing = true;
        try {
          await authStore.refresh();
        } finally {
          this._refreshing = false;
        }
        if (authStore.accessToken) {
          return this._requestInner<T>(method, path, body);
        }
      }
      throw err;
    }
  }

  async createPage(title = 'Untitled'): Promise<Block> {
    return this.request('POST', `${await this.wsPrefix()}/pages`, { title });
  }

  async listPages(): Promise<PageSummary[]> {
    return this.request('GET', `${await this.wsPrefix()}/pages`);
  }

  async getPageTree(id: string): Promise<PageTree> {
    return this.request('GET', `${await this.wsPrefix()}/pages/${id}`);
  }

  async createBlock(parentId: string, type: BlockType, content: any = {}, position?: number): Promise<Block> {
    return this.request('POST', `${await this.wsPrefix()}/blocks`, { parent_id: parentId, type, content, position });
  }

  async updateBlock(id: string, data: { content?: any; type?: BlockType }): Promise<Block> {
    return this.request('PATCH', `${await this.wsPrefix()}/blocks/${id}`, data);
  }

  async deleteBlock(id: string): Promise<void> {
    return this.request('DELETE', `${await this.wsPrefix()}/blocks/${id}`);
  }

  async restoreBlock(id: string): Promise<Block> {
    return this.request('PATCH', `${await this.wsPrefix()}/blocks/${id}/restore`);
  }

  async listTrash(): Promise<PageSummary[]> {
    return this.request('GET', `${await this.wsPrefix()}/trash`);
  }

  async permanentDeleteBlock(id: string): Promise<void> {
    return this.request('DELETE', `${await this.wsPrefix()}/blocks/${id}/permanent`);
  }

  async listFavorites(): Promise<PageSummary[]> {
    return this.request('GET', `${await this.wsPrefix()}/favorites`);
  }

  async toggleFavorite(id: string, favorited: boolean): Promise<Block> {
    return this.request('PATCH', `${await this.wsPrefix()}/blocks/${id}`, { content: { favorited } });
  }

  async moveBlock(id: string, parentId: string | null, position: number): Promise<Block> {
    return this.request('PATCH', `${await this.wsPrefix()}/blocks/${id}/move`, { parent_id: parentId, position });
  }

  updateProfile(data: { name?: string; email?: string; current_password?: string }): Promise<User> {
    return this.request('PATCH', '/auth/me', data);
  }

  updatePassword(data: { current_password: string; new_password: string }): Promise<void> {
    return this.request('PATCH', '/auth/me/password', data);
  }

  deleteAccount(data: { password: string }): Promise<void> {
    return this.request('DELETE', '/auth/me', data);
  }

  async search(query: string, limit = 20, offset = 0): Promise<SearchResult[]> {
    const q = encodeURIComponent(query);
    return this.request('GET', `${await this.wsPrefix()}/search?q=${q}&limit=${limit}&offset=${offset}`);
  }

  async uploadFile(file: File): Promise<{ url: string }> {
    const form = new FormData();
    form.append('file', file);
    const opts: RequestInit = { method: 'POST', body: form };
    if (authStore.accessToken) {
      opts.headers = { 'Authorization': `Bearer ${authStore.accessToken}` };
    }
    const res = await fetch(`${BASE_URL}/uploads`, opts);
    if (!res.ok) {
      const text = await res.text();
      let message = text;
      try { const json = JSON.parse(text); if (json.error) message = json.error; } catch { /* use raw text */ }
      throw new ApiError(res.status, message);
    }
    return res.json();
  }
}

export const api = new ApiClient();
