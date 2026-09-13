import { Router } from 'express';
import { query } from '../db.js';

const router = Router();

const asyncRoute = (fn) => (req, res, next) => fn(req, res, next).catch(next);

// Title is derived from the first markdown heading / first line, gist-style,
// unless the user typed one explicitly.
function deriveTitle(title, content) {
  const explicit = (title || '').trim();
  if (explicit) return explicit.slice(0, 200);
  const firstLine = (content || '').split('\n').find((l) => l.trim());
  if (!firstLine) return 'Untitled';
  return firstLine.replace(/^#+\s*/, '').trim().slice(0, 200) || 'Untitled';
}

// GET /api/notes?q=search
router.get('/', asyncRoute(async (req, res) => {
  const q = (req.query.q || '').trim();
  const { rows } = q
    ? await query(
        `SELECT id, title, content, created_at, updated_at FROM notes
         WHERE title ILIKE $1 OR content ILIKE $1
         ORDER BY updated_at DESC`,
        [`%${q}%`]
      )
    : await query(
        `SELECT id, title, content, created_at, updated_at FROM notes
         ORDER BY updated_at DESC`
      );
  res.json(rows);
}));

router.get('/:id', asyncRoute(async (req, res) => {
  const { rows } = await query('SELECT * FROM notes WHERE id = $1', [req.params.id]);
  if (!rows[0]) return res.status(404).json({ error: 'Note not found' });
  res.json(rows[0]);
}));

router.post('/', asyncRoute(async (req, res) => {
  const { title, content = '' } = req.body || {};
  const { rows } = await query(
    'INSERT INTO notes (title, content) VALUES ($1, $2) RETURNING *',
    [deriveTitle(title, content), content]
  );
  res.status(201).json(rows[0]);
}));

router.put('/:id', asyncRoute(async (req, res) => {
  const { title, content = '' } = req.body || {};
  const { rows } = await query(
    'UPDATE notes SET title = $1, content = $2 WHERE id = $3 RETURNING *',
    [deriveTitle(title, content), content, req.params.id]
  );
  if (!rows[0]) return res.status(404).json({ error: 'Note not found' });
  res.json(rows[0]);
}));

router.delete('/:id', asyncRoute(async (req, res) => {
  const { rowCount } = await query('DELETE FROM notes WHERE id = $1', [req.params.id]);
  if (!rowCount) return res.status(404).json({ error: 'Note not found' });
  res.status(204).end();
}));

export default router;
