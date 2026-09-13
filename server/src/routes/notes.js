import { Router } from 'express';

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

export default function notesRouter(store) {
  const router = Router();

  router.get('/', asyncRoute(async (req, res) => {
    res.json(await store.list((req.query.q || '').trim()));
  }));

  router.get('/:id', asyncRoute(async (req, res) => {
    const note = await store.get(req.params.id);
    if (!note) return res.status(404).json({ error: 'Note not found' });
    res.json(note);
  }));

  router.post('/', asyncRoute(async (req, res) => {
    const { title, content = '' } = req.body || {};
    res.status(201).json(await store.create({ title: deriveTitle(title, content), content }));
  }));

  router.put('/:id', asyncRoute(async (req, res) => {
    const { title, content = '' } = req.body || {};
    const note = await store.update(req.params.id, {
      title: deriveTitle(title, content),
      content,
    });
    if (!note) return res.status(404).json({ error: 'Note not found' });
    res.json(note);
  }));

  router.delete('/:id', asyncRoute(async (req, res) => {
    if (!(await store.remove(req.params.id)))
      return res.status(404).json({ error: 'Note not found' });
    res.status(204).end();
  }));

  return router;
}
