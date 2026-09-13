import path from 'node:path';
import { createSqliteStore } from './sqlite.js';

const WELCOME = `# Welcome

This editor speaks **Markdown**, just like GitHub Gists.

- **Bold** with \`Ctrl+B\`
- *Italic* with \`Ctrl+I\`
- Links with \`Ctrl+K\`

1. Numbered lists work
2. So do task lists:

- [x] Write a note
- [ ] Write another one

> Blockquotes, tables and code fences render too.
`;

// DATABASE_URL present -> Postgres (the Docker stack).
// Otherwise -> a local SQLite file, no services required.
export async function openStore() {
  const url = process.env.DATABASE_URL;

  let store;
  if (url) {
    // Imported lazily: running on SQLite needs no Postgres driver installed.
    const { createPostgresStore } = await import('./postgres.js');
    store = createPostgresStore(url);
  } else {
    const file =
      process.env.SQLITE_PATH ||
      path.resolve(import.meta.dirname, '../../../data/notes.db');
    store = createSqliteStore(file);
  }

  await store.ready();
  await store.seed('Welcome to your notes', WELCOME);
  return store;
}
