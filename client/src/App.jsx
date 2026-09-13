import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import NoteList from '@/components/NoteList';
import Editor from '@/components/Editor';
import SavedView from '@/components/SavedView';
import EmptyState from '@/components/EmptyState';
import { Card } from '@/components/ui/card';
import { TrashDialog } from '@/components/TrashDialog';
import {
  createNote,
  deleteNote,
  emptyTrash,
  listNotes,
  listTrash,
  purgeNote,
  restoreNote,
  updateNote,
} from '@/lib/api';

const blankNote = () => ({ id: null, title: '', content: '', updated_at: null });

export default function App() {
  const [notes, setNotes] = useState([]);
  const [draft, setDraft] = useState(null);     // note open in the editor
  const [viewing, setViewing] = useState(null); // note open read-only
  const [query, setQuery] = useState('');
  const [saving, setSaving] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [trash, setTrash] = useState([]);
  const [trashOpen, setTrashOpen] = useState(false);
  // Ids of deletes, newest last, so undo can walk back through them.
  const [deleted, setDeleted] = useState([]);

  const refresh = useCallback(async (q) => {
    try {
      setNotes(await listNotes(q));
    } catch (e) {
      toast.error('Could not load notes', { description: e.message });
    }
  }, []);

  const refreshTrash = useCallback(async () => {
    try {
      setTrash(await listTrash());
    } catch (e) {
      toast.error('Could not load the trash', { description: e.message });
    }
  }, []);

  useEffect(() => {
    const t = setTimeout(() => refresh(query), 200);
    return () => clearTimeout(t);
  }, [query, refresh]);

  useEffect(() => {
    refreshTrash();
  }, [refreshTrash]);

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
      setDeleted((d) => [...d, id]);
      await Promise.all([refresh(query), refreshTrash()]);
      toast.success('Moved to trash', {
        action: { label: 'Undo', onClick: () => restore(id) },
      });
    } catch (e) {
      toast.error('Delete failed', { description: e.message });
    }
  };

  // Restoring is shared by the undo action and the trash dialog.
  const restore = async (id) => {
    try {
      const note = await restoreNote(id);
      setDeleted((d) => d.filter((x) => x !== id));
      await Promise.all([refresh(query), refreshTrash()]);
      setViewing(note);
      toast.success('Restored');
    } catch (e) {
      toast.error('Restore failed', { description: e.message });
    }
  };

  const undoLastDelete = async () => {
    const id = deleted[deleted.length - 1];
    if (!id) {
      toast('Nothing to undo');
      return;
    }
    await restore(id);
  };

  const purge = async (note) => {
    try {
      await purgeNote(note.id);
      setDeleted((d) => d.filter((x) => x !== note.id));
      await refreshTrash();
      toast.success('Deleted for good');
    } catch (e) {
      toast.error('Could not delete the note', { description: e.message });
    }
  };

  const empty = async () => {
    try {
      const { deleted: n } = await emptyTrash();
      setDeleted([]); // those ids are gone; undo has nothing to return to
      await refreshTrash();
      toast.success(`Emptied the trash (${n} ${n === 1 ? 'note' : 'notes'})`);
    } catch (e) {
      toast.error('Could not empty the trash', { description: e.message });
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
        trashCount={trash.length}
        onOpenTrash={() => setTrashOpen(true)}
        canUndo={deleted.length > 0}
        onUndo={undoLastDelete}
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

      <TrashDialog
        open={trashOpen}
        onOpenChange={setTrashOpen}
        trash={trash}
        onRestore={(note) => restore(note.id)}
        onPurge={purge}
        onEmpty={empty}
      />
    </div>
  );
}
