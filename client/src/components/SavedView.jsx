import { Clock, Pencil, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { DeleteNoteDialog } from '@/components/DeleteNoteDialog';
import { renderMarkdown } from '@/lib/markdown';
import { fullTime, relativeTime } from '@/lib/time';

export default function SavedView({ note, onEdit, onDelete }) {
  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="truncate text-xl font-semibold tracking-tight">{note.title}</h2>
          <div className="text-muted-foreground mt-1 flex items-center gap-1.5 text-xs">
            <Clock className="size-3.5" />
            <span title={fullTime(note.updated_at)}>
              Updated {relativeTime(note.updated_at)}
            </span>
            <Badge variant="outline" className="ml-1">Saved</Badge>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={onEdit}>
            <Pencil /> Edit
          </Button>
          <DeleteNoteDialog title={note.title} onConfirm={() => onDelete(note.id)}>
            <Button variant="outline" size="icon" aria-label="Delete note">
              <Trash2 className="text-destructive" />
            </Button>
          </DeleteNoteDialog>
        </div>
      </div>

      <Separator />

      <article
        className="prose prose-zinc dark:prose-invert min-h-0 max-w-none flex-1 overflow-y-auto pb-6"
        dangerouslySetInnerHTML={{ __html: renderMarkdown(note.content) }}
      />
    </div>
  );
}
