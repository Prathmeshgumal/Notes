# site

Live at **<https://nib-note.vercel.app>**.

The landing page. One static HTML file with no build step — the fonts come from
Google Fonts and GSAP from a CDN, so there is nothing to compile and nothing to
install.

## Look at it locally

```bash
python3 -m http.server -d site 8000
```

Then open <http://localhost:8000>.

## Publish it

```bash
npx vercel --cwd site           # a preview URL, to check
npx vercel --cwd site --prod    # the real thing
```

The first run asks you to sign in and to confirm the project name; everything
after that is one command. Vercel serves the directory as-is — there is no
framework to detect and no build command to set.

`og.png` is the link-preview card. Social platforms do not render SVG, so it is
a PNG, generated from `docs/og-card.html`:

```bash
google-chrome --headless --disable-gpu \
  --window-size=1200,630 --screenshot=site/og.png docs/og-card.html
```

## Editing it

`index.html` holds the markup, the styles and the scripts, in that order. The
terminal frame in the hero is generated from real output rather than typed by
hand — see `docs/README.md` for how, and regenerate it rather than editing the
box-drawing characters, or the borders will stop lining up.
