package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	title     lipgloss.Style
	subtitle  lipgloss.Style
	tab       lipgloss.Style
	tabActive lipgloss.Style
	header    lipgloss.Style
	footer    lipgloss.Style
	label     lipgloss.Style
	value     lipgloss.Style
	muted     lipgloss.Style
	ok        lipgloss.Style
	fail      lipgloss.Style
	warn      lipgloss.Style
	pane      lipgloss.Style
	border    lipgloss.Style
	focus     lipgloss.Style
	barEmpty  lipgloss.Style
	spark     lipgloss.Style
	info      lipgloss.Style
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
			warn:      plain,
			pane:      plain,
			border:    plain,
			focus:     plain,
			barEmpty:  plain,
			spark:     plain,
			info:      plain,
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
		warn: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		pane: lipgloss.NewStyle().
			Padding(0, 1),
		border: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("238")),
		focus: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")),
		barEmpty: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
		spark: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")),
	}
}

func (s styles) severityStyle(severity string) lipgloss.Style {
	switch severity {
	case "critical", "error":
		return s.fail
	case "warning", "warn":
		return s.warn
	default:
		return s.info
	}
}

func (s styles) healthStyle(ok bool) lipgloss.Style {
	if ok {
		return s.ok
	}
	return s.fail
}
