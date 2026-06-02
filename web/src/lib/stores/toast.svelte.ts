import type { Block } from '$lib/types';

let deletedBlock = $state<Block | null>(null);
let timer: ReturnType<typeof setTimeout> | null = null;

export function showUndoToast(block: Block) {
  deletedBlock = block;
  if (timer) clearTimeout(timer);
  timer = setTimeout(() => {
    deletedBlock = null;
  }, 5000);
}

export function clearToast() {
  deletedBlock = null;
  if (timer) clearTimeout(timer);
}

export function getDeletedBlock() {
  return deletedBlock;
}
