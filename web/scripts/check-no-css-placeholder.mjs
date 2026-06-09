import { readFileSync, readdirSync } from 'node:fs';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const srcDir = join(fileURLToPath(new URL('..', import.meta.url)), 'src');

function* svelteFiles(dir) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const fullPath = join(dir, entry.name);
    if (entry.isDirectory()) {
      yield* svelteFiles(fullPath);
    } else if (entry.name.endsWith('.svelte')) {
      yield fullPath;
    }
  }
}

let hasErrors = false;

for (const file of svelteFiles(srcDir)) {
  const content = readFileSync(file, 'utf-8');
  const styleMatch = content.match(/<style[^>]*>([\s\S]*?)<\/style>/);
  if (!styleMatch) continue;

  const css = styleMatch[1];
  const pseudoBlocks = css.matchAll(/[^{]*::(before|after)\s*\{([^}]*)\}/g);
  for (const [, , body] of pseudoBlocks) {
    const m = body.match(/content\s*:\s*(['"])([^'"]*)\1\s*;?\s*/);
    if (m && m[2] !== '') {
      console.error(`FAIL: ${relative(process.cwd(), file)} — ::${m[1]} non-empty content: '${m[2]}'`);
      hasErrors = true;
    }
  }
}

if (hasErrors) {
  process.exit(1);
} else {
  console.log('PASS: No CSS content property placeholders found');
}
