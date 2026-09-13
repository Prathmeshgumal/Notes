import { existsSync } from 'node:fs';
import path from 'node:path';
import express from 'express';
import cors from 'cors';
import notesRouter from './routes/notes.js';
import { openStore } from './store/index.js';

const PORT = Number(process.env.PORT || 4000);
const HOST = process.env.HOST || '127.0.0.1';

const store = await openStore();

const app = express();
app.use(cors());
app.use(express.json({ limit: '1mb' }));

app.get('/api/health', (_req, res) => res.json({ ok: true, store: store.kind }));
app.use('/api/notes', notesRouter(store));

// When the client has been built, serve it from this same process so the whole
// app is one command on one port. In Docker, nginx does this instead.
const clientDir =
  process.env.CLIENT_DIR || path.resolve(import.meta.dirname, '../../client/dist');
const servingUI = existsSync(path.join(clientDir, 'index.html'));
if (servingUI) {
  app.use(express.static(clientDir));
  app.get(/^(?!\/api\/).*/, (_req, res) => res.sendFile(path.join(clientDir, 'index.html')));
}

app.use((req, res) => res.status(404).json({ error: `No route for ${req.method} ${req.path}` }));

// Invalid UUIDs reach pg as 22P02; surface those as 400 rather than 500.
app.use((err, _req, res, _next) => {
  console.error(err);
  if (err.code === '22P02') return res.status(400).json({ error: 'Invalid id' });
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, HOST, () => {
  console.log(`\n  Notes running at http://localhost:${PORT}`);
  console.log(`  Storage: ${store.describe()}`);
  if (!servingUI) console.log('  (API only — run the client dev server for the UI)');
  console.log('  Ctrl+C to stop\n');
});
