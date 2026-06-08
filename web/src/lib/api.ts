import { authStore } from '$lib/stores/auth.svelte';
import type { Block, BlockType, PageSummary, PageTree } from './types';

const BASE_URL = '/api/v1';

export class ApiError extends Error {
  constructor(public status: number, text: string) {
    super(text);
  }
}

class ApiClient {
  private async requestInner<T>(method: string, path: string, body?: any): Promise<T> {
    const opts: RequestInit = { method };
    if (body !== undefined) {
      opts.headers = { 'Content-Type': 'application/json' };
      opts.body = JSON.stringify(body);
    }
    if (authStore.accessToken) {
      opts.headers = { ...(opts.headers as Record<string, string>), 'Authorization': `Bearer ${authStore.accessToken}` };
    }
    const res = await fetch(`${BASE_URL}${path}`, opts);
    if (!res.ok) throw new ApiError(res.status, await res.text());
    if (res.status === 204) return undefined as T;
    return res.json();
  }

  async request<T>(method: string, path: string, body?: any): Promise<T> {
    try {
      return await this.requestInner<T>(method, path, body);
    } catch (err: any) {
      if (err instanceof ApiError && err.status === 401) {
        await authStore.refresh();
        if (authStore.accessToken) {
          return this.requestInner<T>(method, path, body);
        }
      }
      throw err;
    }
  }

  createPage(title = 'Untitled'): Promise<Block> {
    return this.request('POST', '/pages', { title });
  }

  listPages(): Promise<PageSummary[]> {
    return this.request('GET', '/pages');
  }

  getPageTree(id: string): Promise<PageTree> {
    return this.request('GET', `/pages/${id}`);
  }

  createBlock(parentId: string, type: BlockType, content: any = {}, position?: number): Promise<Block> {
    return this.request('POST', '/blocks', { parent_id: parentId, type, content, position });
  }

  updateBlock(id: string, data: { content?: any; type?: BlockType }): Promise<Block> {
    return this.request('PATCH', `/blocks/${id}`, data);
  }

  deleteBlock(id: string): Promise<void> {
    return this.request('DELETE', `/blocks/${id}`);
  }

  restoreBlock(id: string): Promise<Block> {
    return this.request('PATCH', `/blocks/${id}/restore`);
  }

  listFavorites(): Promise<PageSummary[]> {
    return this.request('GET', '/favorites');
  }

  moveBlock(id: string, parentId: string | null, position: number): Promise<Block> {
    return this.request('PATCH', `/blocks/${id}/move`, { parent_id: parentId, position });
  }

  async uploadFile(file: File): Promise<{ url: string }> {
    const form = new FormData();
    form.append('file', file);
    const opts: RequestInit = { method: 'POST', body: form };
    if (authStore.accessToken) {
      opts.headers = { 'Authorization': `Bearer ${authStore.accessToken}` };
    }
    const res = await fetch(`${BASE_URL}/uploads`, opts);
    if (!res.ok) throw new ApiError(res.status, await res.text());
    return res.json();
  }
}

export const api = new ApiClient();
