import type { Block, BlockType, PageSummary, PageTree } from './types';

const BASE_URL = '';

export class ApiError extends Error {
  constructor(public status: number, text: string) {
    super(text);
  }
}

class ApiClient {
  private async request<T>(method: string, path: string, body?: any): Promise<T> {
    const opts: RequestInit = { method };
    if (body !== undefined) {
      opts.headers = { 'Content-Type': 'application/json' };
      opts.body = JSON.stringify(body);
    }
    const res = await fetch(`${BASE_URL}${path}`, opts);
    if (!res.ok) throw new ApiError(res.status, await res.text());
    if (res.status === 204) return undefined as T;
    return res.json();
  }

  createPage(title = 'Untitled'): Promise<Block> {
    return this.request('POST', '/api/v1/pages', { title });
  }

  listPages(): Promise<PageSummary[]> {
    return this.request('GET', '/api/v1/pages');
  }

  getPageTree(id: string): Promise<PageTree> {
    return this.request('GET', `/api/v1/pages/${id}`);
  }

  createBlock(
    parentId: string,
    type: BlockType,
    content: any = {},
    position?: number
  ): Promise<Block> {
    return this.request('POST', '/api/v1/blocks', {
      parent_id: parentId,
      type,
      content,
      position,
    });
  }

  updateBlock(id: string, data: { content?: any; type?: BlockType }): Promise<Block> {
    return this.request('PATCH', `/api/v1/blocks/${id}`, data);
  }

  deleteBlock(id: string): Promise<void> {
    return this.request('DELETE', `/api/v1/blocks/${id}`);
  }

  restoreBlock(id: string): Promise<Block> {
    return this.request('PATCH', `/api/v1/blocks/${id}/restore`);
  }

  moveBlock(id: string, parentId: string | null, position: number): Promise<Block> {
    return this.request('PATCH', `/api/v1/blocks/${id}/move`, {
      parent_id: parentId,
      position,
    });
  }

  async uploadFile(file: File): Promise<{ url: string }> {
    const form = new FormData();
    form.append('file', file);
    const res = await fetch(`${BASE_URL}/api/v1/uploads`, { method: 'POST', body: form });
    if (!res.ok) throw new ApiError(res.status, await res.text());
    return res.json();
  }
}

export const api = new ApiClient();
