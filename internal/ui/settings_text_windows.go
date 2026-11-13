package ui

// Text measurement for the settings window.
//
// Two things live here because they must not drift apart:
//
//   - wrapHint breaks a caption into lines, and the settings window *draws*
//     those lines rather than handing the paragraph to DrawText. If the layout
//     reserved two lines and DrawText decided it needed three, the third would
//     be clipped with no way to see it.
//   - settingsRowsMeasured fills in how many lines each row's hint actually
//     needs at the current width, which is what the row heights are built from.
//
// The wrap is greedy and CJK-aware: a run of characters that contains no spaces
// (Chinese, Japanese) may break between any two characters, while a run of
// Latin text breaks at spaces and hyphens. That is close enough to what GDI's
// DT_WORDBREAK does, and it is exact for the reason that matters: the same
// function decides the reserved height and the drawn lines.

import "strings"

// settingsHintMaxLines is how many lines a caption may take before it is
// ellipsised. Two for a row hint, three for a note, because a note is a
// paragraph with no control beside it.
const (
	settingsHintMaxLines = 2
	settingsNoteMaxLines = 3
)

func settingsMaxLines(kind int) int {
	if kind == rowNote {
		return settingsNoteMaxLines
	}
	return settingsHintMaxLines
}

// wrapHint splits value into at most maxLines lines that each fit width device
// pixels at fontHeight. When it does not fit, the last line is ellipsised.
func wrapHint(target *canvas, value string, fontHeight, weight, width, maxLines int) []string {
	value = strings.TrimSpace(value)
	if value == "" || width <= 0 || maxLines < 1 {
		return nil
	}
	if target.measure(value, fontHeight, weight) <= width {
		return []string{value}
	}

	var lines []string
	current := ""
	// flush appends the current line, and reports whether there is room for
	// another one.
	flush := func() bool {
		if current == "" {
			return false
		}
		lines = append(lines, current)
		current = ""
		return len(lines) < maxLines
	}
	fits := func(candidate string) bool {
		return target.measure(candidate, fontHeight, weight) <= width
	}

	for _, chunk := range splitBreakable(value) {
		candidate := current + chunk
		if fits(candidate) {
			current = candidate
			continue
		}
		// A Latin word may be longer than the column: break inside it rather
		// than losing it.
		if current != "" && !flush() {
			return ellipsise(target, lines, value, fontHeight, weight, width)
		}
		if fits(chunk) {
			current = chunk
			continue
		}
		for _, char := range chunk {
			piece := current + string(char)
			if !fits(piece) {
				if !flush() {
					return ellipsise(target, lines, value, fontHeight, weight, width)
				}
				piece = string(char)
			}
			current = piece
		}
	}
	if current != "" {
		if len(lines) < maxLines {
			lines = append(lines, current)
		} else {
			return ellipsise(target, lines, value, fontHeight, weight, width)
		}
	}
	return pullOrphanPunctuation(lines)
}

// splitBreakable cuts value into chunks that may be joined without changing its
// text: a run of non-space characters followed by its trailing spaces. Every
// boundary between chunks is a legal line break.
func splitBreakable(value string) []string {
	var chunks []string
	start := 0
	for index, char := range value {
		if char == ' ' || char == '\t' {
			// End the chunk after the run of spaces.
			next := index
			for next < len(value) && (value[next] == ' ' || value[next] == '\t') {
				next++
			}
			chunks = append(chunks, value[start:next])
			start = next
		}
	}
	if start < len(value) {
		chunks = append(chunks, value[start:])
	}
	return chunks
}

// lineStartForbidden lists the characters that may not begin a line in Chinese
// typesetting (行首禁则). Without this a wrap can leave a lone "。" on its own
// line, which is what the appearance tab's note did the first time it was
// rendered.
const lineStartForbidden = "。，、；：？！）］｝」』】》〉・…％”’"

// pullOrphanPunctuation moves the last character of a line down when the next
// line would otherwise start with a character that may not start one.
func pullOrphanPunctuation(lines []string) []string {
	for index := 1; index < len(lines); index++ {
		for {
			current := []rune(lines[index])
			if len(current) == 0 {
				break
			}
			forbidden := false
			for _, char := range lineStartForbidden {
				if current[0] == char {
					forbidden = true
					break
				}
			}
			if !forbidden {
				break
			}
			previous := []rune(lines[index-1])
			if len(previous) <= 1 {
				break
			}
			lines[index-1] = string(previous[:len(previous)-1])
			lines[index] = string(previous[len(previous)-1:]) + string(current)
		}
	}
	return lines
}

// ellipsise fills the last allowed line with as much text as fits, ending in an
// ellipsis, so a clipped caption still reads as clipped.
func ellipsise(target *canvas, lines []string, value string, fontHeight, weight, width int) []string {
	if len(lines) == 0 {
		return []string{"…"}
	}
	last := lines[len(lines)-1]
	for len(last) > 0 {
		candidate := strings.TrimRight(last, " ") + "…"
		if target.measure(candidate, fontHeight, weight) <= width {
			lines[len(lines)-1] = candidate
			return lines
		}
		runes := []rune(last)
		last = string(runes[:len(runes)-1])
	}
	lines[len(lines)-1] = "…"
	return lines
}

// measureSettingsRows fills in each row's hint line count for the given width
// and type ramp. It is called once per reflow, never per frame.
func measureSettingsRows(target *canvas, rows []settingsRow, fonts settingsTypeRamp, width int) []settingsRow {
	if target == nil {
		return rows
	}
	measured := make([]settingsRow, len(rows))
	copy(measured, rows)
	for index := range measured {
		row := &measured[index]
		// A note carries its paragraph in title, not hint: it has no control to
		// sit beside, so the title line *is* the text. Measuring hint here left
		// every note reserving one line and drawing two.
		text, font := row.hint, fonts.hint
		if row.kind == rowNote {
			text, font = row.title, fonts.note
		}
		if text == "" {
			row.hintLines = 1
			continue
		}
		lines := wrapHint(target, text, font, 400, width, settingsMaxLines(row.kind))
		if len(lines) == 0 {
			row.hintLines = 1
			continue
		}
		row.hintLines = len(lines)
	}
	return measured
}
