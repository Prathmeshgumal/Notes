CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS notes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title       TEXT NOT NULL DEFAULT 'Untitled',
  content     TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notes_updated_at_idx ON notes (updated_at DESC);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS notes_set_updated_at ON notes;
CREATE TRIGGER notes_set_updated_at
  BEFORE UPDATE ON notes
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO notes (title, content)
SELECT 'Welcome to your notes', E'# Welcome\n\nThis editor speaks **Markdown**, just like GitHub Gists.\n\n- **Bold** with `Ctrl+B`\n- *Italic* with `Ctrl+I`\n- Links with `Ctrl+K` -> [Markdown guide](https://docs.github.com/en/get-started/writing-on-github)\n\n1. Numbered lists work\n2. So do task lists:\n\n- [x] Write a note\n- [ ] Write another one\n\n> Blockquotes, tables and code fences render too.\n\n```js\nconsole.log("hello notes");\n```\n'
WHERE NOT EXISTS (SELECT 1 FROM notes);
