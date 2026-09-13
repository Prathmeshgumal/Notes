const base = '/api';

async function request(path, options = {}) {
  const res = await fetch(base + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed (${res.status})`);
  }
  return res.status === 204 ? null : res.json();
}

export const listNotes = (q = '') =>
  request(`/notes${q ? `?q=${encodeURIComponent(q)}` : ''}`);
export const createNote = (note) =>
  request('/notes', { method: 'POST', body: JSON.stringify(note) });
export const updateNote = (id, note) =>
  request(`/notes/${id}`, { method: 'PUT', body: JSON.stringify(note) });
export const deleteNote = (id) => request(`/notes/${id}`, { method: 'DELETE' });

export const listTrash = () => request('/trash');
export const restoreNote = (id) => request(`/trash/${id}`, { method: 'POST' });
export const purgeNote = (id) => request(`/trash/${id}`, { method: 'DELETE' });
export const emptyTrash = () => request('/trash', { method: 'DELETE' });
