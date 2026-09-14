# docs

The screenshots are real screens, not mock-ups. The terminal UI is run inside a
pseudo-terminal, its output captured, and the ANSI replayed into SVG by
`ansi2svg.py`:

| File | Shows |
| --- | --- |
| `screenshot.svg` | Reading a note — the two-column layout |
| `screenshot-editing.svg` | The editor, with a list carrying itself on |

To regenerate one, run the app under a pty, save the raw output, and pass it
through the converter:

```bash
python3 docs/ansi2svg.py capture.raw docs/screenshot.svg "nib"
```

`COLS, ROWS` at the top of the converter must match the pseudo-terminal's size,
or the replay will wrap in the wrong places.

SVG rather than PNG so the text stays crisp at any size, the file is a few
kilobytes, and a diff is readable.
