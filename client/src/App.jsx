import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import NoteList from '@/components/NoteList';
import Editor from '@/components/Editor';
import SavedView from '@/components/SavedView';
import EmptyState from '@/components/EmptyState';
import { Card } from '@/components/ui/card';
import { createNote, deleteNote, listNotes, updateNote } from '@/lib/api';

const blankNote = () => ({ id: null, title: '', content: '', updated_at: null });

export default function App() {
  const [notes, setNotes] = useState([]);
  const [draft, setDraft] = useState(null);     // note open in the editor
  const [viewing, setViewing] = useState(null); // note open read-only
  const [query, setQuery] = useState('');
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);

  const refresh = useCallback(async (q) => {
    try {
      setNotes(await listNotes(q));
    } catch (e) {
      toast.error('Could not load notes', { description: e.message });
    }
  }, []);

  useEffect(() => {
    const t = setTimeout(() => refresh(query), 200);
    return () => clearTimeout(t);
  }, [query, refresh]);

  const startNew = () => {
    setViewing(null);
    setDraft(blankNote());
    setDirty(false);
  };

  const save = async () => {
    if (!draft) return;
    setSaving(true);
    try {
      const payload = { title: draft.title, content: draft.content };
      const saved = draft.id
        ? await updateNote(draft.id, payload)
        : await createNote(payload);
      setDirty(false);
      setDraft(null);
      setViewing(saved);
      await refresh(query);
      toast.success(draft.id ? 'Note saved' : 'Note created');
    } catch (e) {
      toast.error('Save failed', { description: e.message });
    } finally {
      setSaving(false);
    }
  };

  const remove = async (id) => {
    try {
      await deleteNote(id);
      setDraft(null);
      setViewing(null);
      await refresh(query);
      toast.success('Note deleted');
    } catch (e) {
      toast.error('Delete failed', { description: e.message });
    }
  };

  return (
    <div className="flex h-svh flex-col md:flex-row">
      <NoteList
        notes={notes}
        selectedId={viewing?.id ?? draft?.id}
        onSelect={(note) => {
          setDraft(null);
          setViewing(note);
        }}
        onNew={startNew}
        query={query}
        onQuery={setQuery}
      />

      <main className="flex min-h-0 flex-1 flex-col p-4 md:p-6">
        <Card className="flex min-h-0 flex-1 flex-col gap-0 p-5">
          {draft ? (
            <Editor
              note={draft}
              onChange={(next) => {
                setDraft(next);
                setDirty(true);
              }}
              onSave={save}
              onDelete={remove}
              saving={saving}
              dirty={dirty || !draft.id}
            />
          ) : viewing ? (
            <SavedView
              note={viewing}
              onDelete={remove}
              onEdit={() => {
                setDraft(viewing);
                setViewing(null);
                setDirty(false);
              }}
            />
          ) : (
            <EmptyState onNew={startNew} />
          )}
        </Card>
      </main>
    </div>
  );
}
