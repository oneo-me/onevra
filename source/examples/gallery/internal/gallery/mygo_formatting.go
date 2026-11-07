package gallery

import (
	"time"

	"github.com/egoist/mygo/i18n"
	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/components"
)

// exampleFormatting writes numbers, money and times in the page's language.
func exampleFormatting(c *ui.Context, _ *demoState) {
	t := c.Theme()
	locale := c.Locale()
	start := time.Date(2026, 6, 15, 9, 30, 0, 0, time.Local)
	rows := [][2]string{
		{"Notation", i18n.FormatNumber(locale, 1234567, i18n.NumberOptions{Notation: i18n.CompactShort})},
		{"CompactLong", i18n.FormatNumber(locale, 1234567, i18n.NumberOptions{Notation: i18n.CompactLong})},
		{"Currency", i18n.FormatNumber(locale, 1234.5, i18n.NumberOptions{Currency: i18n.CurrencyOf("EUR")})},
		{"SignDisplay", i18n.FormatNumber(locale, .034, i18n.NumberOptions{Percent: true, MaxFractionDigits: 1, SignDisplay: i18n.SignExceptZero})},
		{"FormatDateRange", i18n.FormatDateRange(locale, start, start.AddDate(0, 0, 6), i18n.DateOptions{Date: i18n.MediumStyle})},
		{"FormatRelativeTime", i18n.FormatRelativeTime(locale, -3, i18n.RelativeMinute, i18n.RelativeTimeOptions{})},
		{"RelativeShort", i18n.FormatRelativeTime(locale, 2, i18n.RelativeWeek, i18n.RelativeTimeOptions{Style: i18n.RelativeShort})},
		{"Auto", i18n.FormatRelativeTime(locale, -1, i18n.RelativeDay, i18n.RelativeTimeOptions{Auto: true})},
	}
	ui.Column(c).Gap(t.Space(2)).Children(func() {
		for _, row := range rows {
			ui.Row(c).Key(row[0]).Gap(t.Space(3)).Children(func() {
				components.Text(c, row[0]).Width(t.Space(44)).Font("monospace").TextColor(t.TextMuted)
				components.Text(c, row[1])
			})
		}
	})
}
