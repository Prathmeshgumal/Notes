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

Each row is emitted as a single `<text>` carrying an explicit x for every
column, so a cell can never drift from the grid. Positioning colour runs
individually instead lets rounding accumulate across a row, which pulls the
box-drawing verticals out of line and splits the borders.

Capture the terminal at a height where the frame you want is as tall as the one
it replaced. The app repaints only the rows that changed, so a shorter frame
leaves rows of the previous screen underneath it.

SVG rather than PNG so the text stays crisp at any size, the file is a few
kilobytes, and a diff is readable.
