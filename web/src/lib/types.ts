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
  | 'image';

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

export interface PageSummary {
  id: string;
  title: string;
  created_at: string;
}
