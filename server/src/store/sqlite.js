import { DatabaseSync } from 'node:sqlite';
import { randomUUID } from 'node:crypto';
import { mkdirSync } from 'node:fs';
import { dirname } from 'node:path';

// Zero-dependency storage: node:sqlite is built into Node 22.5+.
export function createSqliteStore(file) {
  mkdirSync(dirname(file), { recursive: true });
  const db = new DatabaseSync(file);

  db.exec(`
    PRAGMA journal_mode = WAL;
    CREATE TABLE IF NOT EXISTS notes (
      id         TEXT PRIMARY KEY,
      title      TEXT NOT NULL DEFAULT 'Untitled',
      content    TEXT NOT NULL DEFAULT '',
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );
    CREATE INDEX IF NOT EXISTS notes_updated_at_idx ON notes (updated_at DESC);
  `);

  const now = () => new Date().toISOString();

  return {
    kind: 'sqlite',
    describe: () => `SQLite (${file})`,
    async ready() {},

    async list(q) {
      if (!q) {
        return db
          .prepare('SELECT * FROM notes ORDER BY updated_at DESC')
          .all();
      }
      // SQLite's LIKE is already case-insensitive for ASCII.
      return db
        .prepare(
          `SELECT * FROM notes
           WHERE title LIKE ? OR content LIKE ?
           ORDER BY updated_at DESC`
        )
        .all(`%${q}%`, `%${q}%`);
    },

    async get(id) {
      return db.prepare('SELECT * FROM notes WHERE id = ?').get(id) ?? null;
    },

    async create({ title, content }) {
      const row = {
        id: randomUUID(),
        title,
        content,
        created_at: now(),
        updated_at: now(),
      };
      db.prepare(
        `INSERT INTO notes (id, title, content, created_at, updated_at)
         VALUES (?, ?, ?, ?, ?)`
      ).run(row.id, row.title, row.content, row.created_at, row.updated_at);
      return row;
    },

    async update(id, { title, content }) {
      const res = db
        .prepare('UPDATE notes SET title = ?, content = ?, updated_at = ? WHERE id = ?')
        .run(title, content, now(), id);
      return res.changes ? this.get(id) : null;
    },

    async remove(id) {
      return db.prepare('DELETE FROM notes WHERE id = ?').run(id).changes > 0;
    },

    async count() {
      return db.prepare('SELECT count(*) AS n FROM notes').get().n;
    },

    async seed(title, content) {
      if ((await this.count()) === 0) await this.create({ title, content });
    },
  };
}
