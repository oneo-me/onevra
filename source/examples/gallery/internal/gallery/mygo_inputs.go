package gallery

import (
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// exampleDecimalInput edits an exact amount kept as a string, which typing
// never rounds: too many decimals are marked as an error instead.
func exampleDecimalInput(c *ui.Context, s *demoState) {
	ui.DecimalInput(c, &s.text, "0", "10000", "0.01").Label(tr(c, "decimalAmount")).Width(c.Theme().Space(50))
	components.Text(c, s.text)
}

// exampleInputStyling styles Onevra text boxes with MyGo's input methods: a
// tinted hint, a block caret and a link marked with dots.
func exampleInputStyling(c *ui.Context, s *demoState) {
	t := c.Theme()
	components.TextBox(c, &s.text, components.TextBoxOptions{}).Label(tr(c, "inputConfirmDelete")).
		Placeholder(tr(c, "inputTypeDelete")).PlaceholderColor(t.Danger).FillWidth().MaxWidth(t.Space(80))
	components.TextBox(c, &s.query, components.TextBoxOptions{}).Label(tr(c, "inputBlockCaret")).
		Placeholder(tr(c, "inputBlockCaret")).CaretWidth(t.FontSize * .6).Font("monospace").FillWidth().MaxWidth(t.Space(80))
	components.Text(c, tr(c, "inputDottedLink")).TextColor(t.Accent).DottedUnderline()
}
