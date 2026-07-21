package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	title    lipgloss.Style
	subtitle lipgloss.Style
	tab      lipgloss.Style
	tabActive lipgloss.Style
	header   lipgloss.Style
	footer   lipgloss.Style
	label    lipgloss.Style
	value    lipgloss.Style
	muted    lipgloss.Style
	ok       lipgloss.Style
	fail     lipgloss.Style
	pane     lipgloss.Style
	border   lipgloss.Style
}

func newStyles(noColor bool) styles {
	if noColor {
		plain := lipgloss.NewStyle()
		return styles{
			title:     plain,
			subtitle:  plain,
			tab:       plain,
			tabActive: plain,
			header:    plain,
			footer:    plain,
			label:     plain,
			value:     plain,
			muted:     plain,
			ok:        plain,
			fail:      plain,
			pane:      plain,
			border:    plain,
		}
	}

	return styles{
		title: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
		subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		tab: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1),
		tabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Padding(0, 1),
		header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		value: lipgloss.NewStyle(),
		muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		ok: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")),
		fail: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		pane: lipgloss.NewStyle().
			Padding(0, 1),
		border: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")),
	}
}
