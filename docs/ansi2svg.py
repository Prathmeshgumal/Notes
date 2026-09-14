"""Render a terminal screen, captured from a real pty, as an SVG.

The screen is replayed into a grid so cursor movement and redraws land where
they would on screen, then each run of same-coloured cells becomes one <tspan>.
"""
import re, sys, html

COLS, ROWS = 104, 16

# xterm-256 palette
def xterm256(n):
    if n < 16:
        base = [(0,0,0),(205,0,0),(0,205,0),(205,205,0),(0,0,238),(205,0,205),
                (0,205,205),(229,229,229),(127,127,127),(255,0,0),(0,255,0),
                (255,255,0),(92,92,255),(255,0,255),(0,255,255),(255,255,255)]
        return base[n]
    if n < 232:
        n -= 16
        levels = [0,95,135,175,215,255]
        return (levels[n//36], levels[(n//6)%6], levels[n%6])
    v = 8 + (n-232)*10
    return (v,v,v)

class Screen:
    def __init__(self):
        self.grid = [[(' ', None, False) for _ in range(COLS)] for _ in range(ROWS)]
        self.x = self.y = 0
        self.fg = None
        self.bold = False

    def put(self, ch):
        if self.x >= COLS:
            self.x = 0; self.y += 1
        if 0 <= self.y < ROWS and 0 <= self.x < COLS:
            self.grid[self.y][self.x] = (ch, self.fg, self.bold)
        self.x += 1

    def sgr(self, params):
        ps = [int(p) for p in params.split(';') if p != ''] or [0]
        i = 0
        while i < len(ps):
            p = ps[i]
            if p == 0: self.fg, self.bold = None, False
            elif p == 1: self.bold = True
            elif p == 22: self.bold = False
            elif p == 39: self.fg = None
            elif 30 <= p <= 37: self.fg = xterm256(p-30)
            elif 90 <= p <= 97: self.fg = xterm256(p-90+8)
            elif p == 38 and i+2 < len(ps) and ps[i+1] == 5:
                self.fg = xterm256(ps[i+2]); i += 2
            elif p == 38 and i+4 < len(ps) and ps[i+1] == 2:
                self.fg = (ps[i+2], ps[i+3], ps[i+4]); i += 4   # 24-bit colour
            i += 1

def replay(data):
    s = Screen()
    i = 0
    while i < len(data):
        c = data[i]
        if c == '\x1b':
            m = re.match(r'\x1b\[([0-9;?]*)([a-zA-Z])', data[i:])
            if m:
                params, cmd = m.group(1), m.group(2)
                if cmd == 'm': s.sgr(params)
                elif cmd == 'H':
                    p = [int(x) for x in params.split(';') if x] or [1,1]
                    s.y = (p[0] if len(p)>0 else 1)-1
                    s.x = (p[1] if len(p)>1 else 1)-1
                elif cmd == 'J':
                    if params in ('2', '3'):
                        s.grid = [[(' ',None,False) for _ in range(COLS)] for _ in range(ROWS)]
                        s.x = s.y = 0
                    elif params in ('', '0'):
                        # erase from the cursor to the end of the screen, which
                        # is how a shorter frame clears the taller one below it
                        for xx in range(s.x, COLS):
                            if 0 <= s.y < ROWS: s.grid[s.y][xx] = (' ', None, False)
                        for yy in range(s.y + 1, ROWS):
                            s.grid[yy] = [(' ', None, False) for _ in range(COLS)]
                elif cmd == 'K':
                    for xx in range(s.x, COLS): s.grid[s.y][xx] = (' ',None,False)
                elif cmd == 'A': s.y -= int(params or 1)
                elif cmd == 'B': s.y += int(params or 1)
                elif cmd == 'C': s.x += int(params or 1)
                elif cmd == 'D': s.x -= int(params or 1)
                i += m.end(); continue
            m = re.match(r'\x1b\][^\x07\x1b]*(\x07|\x1b\\)', data[i:])
            if m: i += m.end(); continue
            i += 1; continue
        if c == '\n': s.y += 1; s.x = 0; i += 1; continue
        if c == '\r': s.x = 0; i += 1; continue
        if c == '\x08': s.x = max(0, s.x-1); i += 1; continue
        if ord(c) < 32: i += 1; continue
        s.put(c); i += 1
    return s.grid

CW, CH = 8.5, 19.0          # cell width / line height
PAD, TOP = 18, 34           # padding, and room for the title bar
BG, FGD = "#17101a", "#cfc2d4"

def to_svg(grid, title="nib"):
    """One <text> per row, with every cell placed on an exact grid.

    Earlier versions emitted a <text> per colour run and positioned each by x.
    Any rounding between the assumed cell width and the browser's actual glyph
    advance then accumulated across a row, which pulled box-drawing verticals
    out of line and split the borders. Here each row is a single <text> whose
    per-character positions are given explicitly, so a cell can never drift.
    """
    w = COLS * CW + PAD * 2
    h = ROWS * CH + TOP + PAD
    out = [
      f'<svg xmlns="http://www.w3.org/2000/svg" width="{w:.0f}" height="{h:.0f}" '
      f'viewBox="0 0 {w:.0f} {h:.0f}" font-family="ui-monospace,SFMono-Regular,'
      f'Menlo,Consolas,&quot;DejaVu Sans Mono&quot;,monospace" font-size="13">',
      f'<rect width="{w:.0f}" height="{h:.0f}" rx="10" fill="{BG}"/>',
      '<circle cx="26" cy="19" r="5.5" fill="#ff5f57"/>',
      '<circle cx="45" cy="19" r="5.5" fill="#febc2e"/>',
      '<circle cx="64" cy="19" r="5.5" fill="#28c840"/>',
      f'<text x="{w/2:.0f}" y="23" fill="#7d6f84" font-size="12" '
      f'text-anchor="middle">{html.escape(title)}</text>',
    ]

    # x for every column, computed once so all rows share the same grid
    xs = ' '.join(f'{PAD + c*CW:.1f}' for c in range(COLS))

    for r, row in enumerate(grid):
        y = TOP + r*CH + 13
        if not any(ch.strip() for ch, _, _ in row):
            continue

        # Split the row into colour runs, but keep them inside one <text> as
        # <tspan>s so the shared x list keeps every cell aligned.
        spans, cur_key, buf = [], None, []
        for ch, fg, bold in row:
            key = (fg, bold)
            if key != cur_key and buf:
                spans.append((cur_key, ''.join(buf)))
                buf = []
            cur_key = key
            buf.append(ch)
        if buf:
            spans.append((cur_key, ''.join(buf)))

        # drop trailing blank run so a row does not pad to full width
        while spans and not spans[-1][1].strip():
            spans.pop()
        if not spans:
            continue

        parts = []
        for (fg, bold), text in spans:
            colour = '#%02x%02x%02x' % fg if fg else FGD
            weight = ' font-weight="600"' if bold else ''
            parts.append(f'<tspan fill="{colour}"{weight}>{html.escape(text)}</tspan>')

        out.append(f'<text x="{xs}" y="{y:.0f}" xml:space="preserve">'
                   + ''.join(parts) + '</text>')

    out.append('</svg>')
    return '\n'.join(out)

if __name__ == '__main__':
    raw = open(sys.argv[1], encoding='utf8', errors='replace').read()
    open(sys.argv[2], 'w').write(to_svg(replay(raw), sys.argv[3] if len(sys.argv) > 3 else 'nib'))
    print(f"  wrote {sys.argv[2]}")
