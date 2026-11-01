package gallery

import (
	"time"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/richtext"
	"github.com/oneo-me/onevra/source/components"
)

// Editable and streamed fixture content stays in English, as other sample
// documents do.
const richTextSample = "# Release notes\n\nEdit **bold**, *italic* and `code` text, or type `- ` to start a list.\n\nA book indents every paragraph after the first, while a screen sets paragraphs apart with room between them.\n\n- [x] Draft the summary\n- [ ] Review the screenshots\n"

const markdownReply = "## Streaming replies\n\nMarkdown arrives **a few words at a time**, and the unfinished tail already shows as what it will become.\n\n- Lists, *emphasis* and `code` render as they stream\n- Finished blocks are never parsed again\n\n```go\nfmt.Println(\"Hello, Onevra\")\n```\n"

func richTextDocument(s *demoState) *richtext.Document {
	if s.richText == nil {
		s.richText = richtext.NewDocument(richtext.ParseMarkdown(richTextSample))
	}
	return s.richText
}

// richTextToolbar formats the editor's selection from the titlebar.
func richTextToolbar(c *ui.Context, s *demoState) {
	doc := richTextDocument(s)
	f := doc.Formatting()
	for _, format := range []struct {
		label string
		state richtext.FormatState
		apply func()
	}{
		{"bold", f.Bold, doc.ToggleBold},
		{"italic", f.Italic, doc.ToggleItalic},
		{"code", f.Code, doc.ToggleCode},
	} {
		on := format.state == richtext.FormatOn
		// Keep the editor's selection while formatting it.
		if components.ToggleButton(c, &on, tr(c, format.label)).KeepFocus().Changed() {
			format.apply()
		}
	}
	components.ToggleButton(c, &s.on, tr(c, "richTextBookLayout")).KeepFocus()
}

// exampleRichTextEditor edits attributed text in page memory only, filling
// the pane with equal room on every side.
func exampleRichTextEditor(c *ui.Context, s *demoState) {
	t := c.Theme()
	room := t.Space(6)
	// Book layout indents paragraphs in place of the room between them.
	indent, spacing := float32(0), float32(.75)
	if s.on {
		indent, spacing = 1.5, 0
	}
	richtext.Editor(c, richTextDocument(s), richtext.EditorOptions{
		Label: tr(c, "richTextLabel"),
		Theme: &richtext.Theme{Padding: [2]float32{room, room}, FirstLineIndent: indent, ParagraphSpacing: spacing},
	}).Fill().Radius(0).Background(ui.Transparent).Border(0, ui.Transparent)
}

// markdownStreamToolbar replays the reply from the titlebar.
func markdownStreamToolbar(c *ui.Context, s *demoState) {
	if components.Button(c, tr(c, "markdownStreamReplay"), components.ButtonOutline).Clicked() {
		s.count++
	}
}

// exampleMarkdownStream reveals a canned reply on the frame clock, so it
// needs no goroutine and settles at once when the desktop reduces motion.
func exampleMarkdownStream(c *ui.Context, s *demoState) {
	reply := ui.Column(c).Key("reply").FillWidth()
	tl := reply.Timeline("stream", ui.TimelineOptions{Run: uint64(s.count)})
	n := int(tl.Tween(ui.Tween{From: 0, To: float32(len(markdownReply)), Duration: 4 * time.Second, Ease: ui.Linear}))
	text := markdownReply[:n]
	for !utf8.ValidString(text) {
		text = text[:len(text)-1]
	}
	reply.Label(tr(c, "markdownStreamLabel")).Children(func() {
		if tl.Settled() {
			markdown.View(c, markdownReply, &s.markdown)
		} else {
			markdown.Stream(c, text, &s.markdown)
		}
	})
}
