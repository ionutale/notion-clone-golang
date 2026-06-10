export type BlockType =
  | 'page'
  | 'text'
  | 'heading_1'
  | 'heading_2'
  | 'heading_3'
  | 'bullet_list_item'
  | 'numbered_list_item'
  | 'toggle'
  | 'divider'
  | 'image'
  | 'code';

export interface Block {
  id: string;
  workspace_id: string;
  parent_id: string | null;
  type: BlockType;
  content: Record<string, any>;
  position: number;
  path: string | null;
  created_by: string | null;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
}

export interface PageTree {
  page: Block;
  blocks: Block[];
}

export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface PageSummary {
  id: string;
  title: string;
  icon?: string | null;
  icon_type?: string | null;
  position: number;
  deleted_at?: string | null;
  created_at: string;
  updated_at?: string;
}

export interface SearchResult {
  block_id: string;
  page_id: string;
  page_title: string;
  block_type: string;
  excerpt: string;
  rank: number;
}

export interface PageCursorResponse {
  items: PageSummary[];
  next_cursor?: number;
  has_more: boolean;
}

export interface TrashCursorResponse {
  items: PageSummary[];
  next_cursor?: string;
  has_more: boolean;
}
