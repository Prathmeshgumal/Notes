package tui

import (
	"regexp"
	"strings"
)

var (
	// [text](https://example.com "optional title") and the image form.
	inlineLink = regexp.MustCompile(`(!?\[[^\]]*\])\(\s*([^()\s]+)(\s+"[^"]*")?\s*\)`)
	// A URL written on its own, which the renderer turns into a link anyway.
	bareURL = regexp.MustCompile(`https?://[^\s<>()\[\]"'\x00-\x1f]+`)
)

// eachLineOutsideCode applies fn to every line that is not inside a fenced
// code block, leaving code samples untouched.
func eachLineOutsideCode(md string, fn func(line string) string) string {
	lines := strings.Split(md, "\n")
	inFence := false
	for i, line := range lines {
		if codeFence.MatchString(line) {
			inFence = !inFence
			continue
		}
		if !inFence {
			lines[i] = fn(line)
		}
	}
	return strings.Join(lines, "\n")
}

// hideLinkTargets rewrites link destinations to a bare anchor so the preview
// shows only the link text, the way a browser does.
//
// The renderer prints the destination after the text, and its style cannot
// suppress that: the separating space is written before the style is applied.
// It does skip anchor-only destinations entirely, so pointing links at "#"
// hides both the URL and the space while keeping the link's own styling. The
// real destinations are read from the note itself by Links.
func hideLinkTargets(md string) string {
	return eachLineOutsideCode(md, func(line string) string {
		return inlineLink.ReplaceAllString(line, "$1(#)")
	})
}

// OrderedTargets returns the destination of every inline link in a note, in the
// order the links appear. Images are excluded because the renderer styles them
// differently, and entries line up one-to-one with the markers in the rendered
// output so each run of link text can be paired with its destination.
func OrderedTargets(md string) []string {
	var out []string
	eachLineOutsideCode(md, func(line string) string {
		for _, m := range inlineLink.FindAllStringSubmatch(line, -1) {
			if strings.HasPrefix(m[1], "!") {
				continue // an image, not a link
			}
			out = append(out, m[2])
		}
		return line
	})
	return out
}

// Links returns every destination in a note, in the order they appear, so the
// reader can open them without seeing the URLs.
func Links(md string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(u string) {
		u = strings.TrimRight(u, ".,;:!?")
		if u == "" || strings.HasPrefix(u, "#") || seen[u] {
			return
		}
		seen[u] = true
		out = append(out, u)
	}

	eachLineOutsideCode(md, func(line string) string {
		rest := line
		for _, m := range inlineLink.FindAllStringSubmatch(line, -1) {
			add(m[2])
			rest = strings.Replace(rest, m[0], "", 1)
		}
		// Whatever is left may still hold a URL written on its own.
		for _, u := range bareURL.FindAllString(rest, -1) {
			add(u)
		}
		return line
	})
	return out
}
