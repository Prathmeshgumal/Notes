import express from 'express';
import cors from 'cors';
import notes from './routes/notes.js';
import { waitForDb } from './db.js';

const app = express();
const PORT = process.env.PORT || 4000;

app.use(cors());
app.use(express.json({ limit: '1mb' }));

app.get('/api/health', (_req, res) => res.json({ ok: true }));
app.use('/api/notes', notes);

app.use((req, res) => res.status(404).json({ error: `No route for ${req.method} ${req.path}` }));

// Invalid UUIDs reach pg as 22P02; surface those as 400 rather than 500.
app.use((err, _req, res, _next) => {
  console.error(err);
  if (err.code === '22P02') return res.status(400).json({ error: 'Invalid id' });
  res.status(500).json({ error: 'Internal server error' });
});

await waitForDb();
app.listen(PORT, () => console.log(`API listening on :${PORT}`));
