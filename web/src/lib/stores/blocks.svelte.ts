import { api } from '$lib/api';
import type { Block, BlockType } from '$lib/types';
import { SvelteMap, SvelteSet } from 'svelte/reactivity';

class BlockStore {
  blocks = $state<SvelteMap<string, Block>>(new SvelteMap());
  pageId = $state<string | null>(null);
  pageTitle = $state<string>('');
  pageIcon = $derived(this.blocks.get(this.pageId ?? '')?.content?.icon ?? null);
  pageIconType = $derived(this.blocks.get(this.pageId ?? '')?.content?.icon_type ?? null);
  pageCover = $derived(this.blocks.get(this.pageId ?? '')?.content?.cover ?? null);
  pageCoverType = $derived(this.blocks.get(this.pageId ?? '')?.content?.cover_type ?? 'color');
  pageCoverColor = $derived(this.blocks.get(this.pageId ?? '')?.content?.cover_color ?? '#e5e7eb');
  loading = $state(false);
  error = $state<string | null>(null);
  favoriteIds = $state<SvelteSet<string>>(new SvelteSet());

  childrenMap = $derived.by(() => {
    const map = new SvelteMap<string | null, string[]>();
    for (const block of this.blocks.values()) {
      const pid = block.parent_id;
      if (!map.has(pid)) map.set(pid, []);
      map.get(pid)!.push(block.id);
    }
    for (const [, ids] of map) {
      ids.sort((a, b) => {
        const ba = this.blocks.get(a)!;
        const bb = this.blocks.get(b)!;
        return ba.position - bb.position;
      });
    }
    return map;
  });

  rootBlocks = $derived(this.childrenMap.get(null) ?? []);

  async loadPage(id: string) {
    this.loading = true;
    this.error = null;
    try {
      const { page, blocks } = await api.getPageTree(id);
      this.pageId = id;
      this.pageTitle = page.content?.title ?? 'Untitled';
      const map = new SvelteMap<string, Block>();
      map.set(page.id, page);
      for (const b of blocks) map.set(b.id, b);
      this.blocks = map;
    } catch (e: any) {
      this.error = e.message ?? 'Failed to load page';
    } finally {
      this.loading = false;
    }
  }

  async createBlock(
    parentId: string | null,
    type: BlockType,
    content: any = {},
    position?: number
  ): Promise<Block> {
    const block = await api.createBlock(parentId ?? this.pageId!, type, content, position);
    this.blocks.set(block.id, block);
    return block;
  }

  async updateBlock(id: string, data: { content?: any; type?: BlockType }): Promise<Block> {
    const updated = await api.updateBlock(id, data);
    this.blocks.set(id, updated);
    return updated;
  }

  async updateCover(cover: string | null, coverType: string, coverColor?: string): Promise<Block | undefined> {
    const block = this.blocks.get(this.pageId ?? '');
    if (!block) return;
    const content: Record<string, unknown> = { ...block.content, cover, cover_type: coverType };
    if (coverColor) content.cover_color = coverColor;
    if (cover === null) {
      delete content.cover;
      delete content.cover_type;
      delete content.cover_color;
    }
    return this.updateBlock(this.pageId!, { content });
  }

  async updateIcon(icon: string | null, iconType: string | null): Promise<Block> {
    const block = this.blocks.get(this.pageId ?? '');
    if (!block) throw new Error('Page block not found');
    const content: Record<string, unknown> = { ...block.content, icon, icon_type: iconType };
    if (icon === null) {
      delete content.icon;
      delete content.icon_type;
    }
    return this.updateBlock(this.pageId!, { content });
  }

  async deleteBlock(id: string): Promise<Block> {
    const block = this.blocks.get(id)!;
    await api.deleteBlock(id);
    this.blocks.delete(id);
    return block;
  }

  async restoreBlock(id: string) {
    const restored = await api.restoreBlock(id);
    this.blocks.set(id, restored);
  }

  async moveBlock(id: string, parentId: string | null, position: number) {
    const moved = await api.moveBlock(id, parentId, position);
    this.blocks.set(id, moved);
  }

  createPage(title = 'Untitled'): Promise<Block> {
    return api.createPage(title);
  }

  async listAllPages() {
    return api.listAllPages();
  }

  async loadFavorites() {
    const pages = await api.listAllFavorites();
    this.favoriteIds = new SvelteSet(pages.map(p => p.id));
  }

  async toggleFavorite(blockId: string) {
    const block = this.blocks.get(blockId);
    const currentlyFavorited = this.favoriteIds.has(blockId);
    // Optimistic update
    if (currentlyFavorited) {
      this.favoriteIds = new SvelteSet([...this.favoriteIds].filter(id => id !== blockId));
    } else {
      this.favoriteIds = new SvelteSet([...this.favoriteIds, blockId]);
    }
    if (block) {
      await this.updateBlock(blockId, { content: { ...block.content, favorited: !currentlyFavorited } });
    } else {
      await api.toggleFavorite(blockId, !currentlyFavorited);
    }
  }

  clear() {
    this.blocks = new SvelteMap();
    this.favoriteIds = new SvelteSet();
    this.pageId = null;
    this.pageTitle = '';
    this.loading = false;
    this.error = null;
  }
}
}

export const blockStore = new BlockStore();
