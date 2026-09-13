# docs

`screenshot.svg` is a real screen, not a mock-up: the terminal UI was run in a
pseudo-terminal, its output captured, and the ANSI replayed into SVG by
`ansi2svg.py`. Regenerate it by running the app under a pty, saving the raw
output, and passing it through:

```bash
python3 docs/ansi2svg.py capture.raw docs/screenshot.svg "note"
```

SVG rather than PNG so it stays crisp on any display and is a few kilobytes.
