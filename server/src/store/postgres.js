import pg from 'pg';

export function createPostgresStore(connectionString) {
  const pool = new pg.Pool({ connectionString, max: 5 });
  const q = (text, params) => pool.query(text, params);

  return {
    kind: 'postgres',
    describe: () => 'PostgreSQL',

    // Postgres may still be booting when the API starts; retry before giving up.
    async ready(attempts = 30, delayMs = 1000) {
      for (let i = 1; i <= attempts; i++) {
        try {
          await q('SELECT 1');
          return;
        } catch (err) {
          if (i === attempts) throw err;
          console.log(`db not ready (${err.code || err.message}), retry ${i}/${attempts}`);
          await new Promise((r) => setTimeout(r, delayMs));
        }
      }
    },

    async list(search) {
      const { rows } = search
        ? await q(
            `SELECT * FROM notes
             WHERE title ILIKE $1 OR content ILIKE $1
             ORDER BY updated_at DESC`,
            [`%${search}%`]
          )
        : await q('SELECT * FROM notes ORDER BY updated_at DESC');
      return rows;
    },

    async get(id) {
      const { rows } = await q('SELECT * FROM notes WHERE id = $1', [id]);
      return rows[0] ?? null;
    },

    async create({ title, content }) {
      const { rows } = await q(
        'INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING *',
        [title, content]
      );
      return rows[0];
    },

    async update(id, { title, content }) {
      const { rows } = await q(
        'UPDATE notes SET title = $1, content = $2 WHERE id = $3 RETURNING *',
        [title, content, id]
      );
      return rows[0] ?? null;
    },

    async remove(id) {
      const { rowCount } = await q('DELETE FROM notes WHERE id = $1', [id]);
      return rowCount > 0;
    },

    // The schema and seed row are created by db/init.sql in the Docker stack.
    async seed() {},
  };
}
