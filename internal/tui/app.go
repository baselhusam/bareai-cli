package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/baselhusam/bareai-cli/internal/probe"
	"github.com/baselhusam/bareai-cli/internal/render"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

const (
	headerLines = 2
	tabBarLines = 1
	footerLines = 1
)

// Tab identifies the active dashboard panel.
type Tab int

const (
	TabOverview Tab = iota
	TabGPU
	TabLLM
	TabDocker
	TabDatabase
	TabProbe
)

const tabCount = 6

func (t Tab) String() string {
	switch t {
	case TabOverview:
		return "Overview"
	case TabGPU:
		return "GPUs"
	case TabLLM:
		return "LLMs"
	case TabDocker:
		return "Docker"
	case TabDatabase:
		return "DBs"
	case TabProbe:
		return "Probe"
	default:
		return "?"
	}
}

type listItem struct {
	title  string
	filter string
	index  int
}

func (i listItem) Title() string       { return i.title }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string {
	if i.filter != "" {
		return i.filter
	}
	return i.title
}

// Model is the root Bubble Tea model.
type Model struct {
	opts    Options
	ctx     context.Context
	styles  styles
	tab     Tab
	width   int
	height  int
	ready   bool
	loading bool
	help    bool
	probing bool

	gen uint64

	snap *snapshot.Snapshot
	history metricHistory

	gpuList    list.Model
	llmList    list.Model
	dockerList list.Model
	dbList     list.Model

	gpuDetail    viewport.Model
	llmDetail    viewport.Model
	dockerDetail viewport.Model
	dbDetail     viewport.Model
	probeVP      viewport.Model

	probeResult   *snapshot.ProbeResult
	probeGen      uint64
	lastProbeIdx  int
	focusDetail   bool
	overviewFocus overviewFocus
	dockerShowAll bool

	spinner spinner.Model
}

func newModel(ctx context.Context, opts Options) Model {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	listWidth := 40
	listHeight := 20

	newList := func() list.Model {
		l := list.New([]list.Item{}, delegate, listWidth, listHeight)
		l.SetShowTitle(false)
		l.SetShowStatusBar(true)
		l.SetShowHelp(false)
		l.SetFilteringEnabled(true)
		l.DisableQuitKeybindings()
		return l
	}

	s := spinner.New()
	s.Spinner = spinner.Dot

	m := Model{
		opts:          opts,
		ctx:           ctx,
		styles:        newStyles(opts.NoColor),
		tab:           TabOverview,
		width:         80,
		height:        24,
		loading:       true,
		gen:           1,
		history:       newMetricHistory(defaultHistoryMax),
		gpuList:       newList(),
		llmList:       newList(),
		dockerList:    newList(),
		dbList:        newList(),
		spinner:       s,
		lastProbeIdx:  -1,
		overviewFocus: overviewFocus{section: sectionHost, row: 0},
		probeVP:       viewport.New(78, 18),
		gpuDetail:     viewport.New(38, 18),
		llmDetail:     viewport.New(38, 18),
		dockerDetail:  viewport.New(38, 18),
		dbDetail:      viewport.New(38, 18),
	}
	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		collectSnapshotCmd(m.ctx, m.opts.Timeout, m.gen, true),
		tickCmd(m.opts.Refresh),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" || key == "q" {
			return m, tea.Quit
		}
		if key == "?" {
			m.help = !m.help
			return m, nil
		}
		if key == "r" {
			m.startRefresh()
			return m, collectSnapshotCmd(m.ctx, m.opts.Timeout, m.gen, true)
		}
		if key == "a" && m.tab == TabDocker {
			m.dockerShowAll = !m.dockerShowAll
			m.syncLists()
			m.syncViewports()
			return m, nil
		}
		if key == "p" && (m.tab == TabLLM || m.tab == TabProbe) {
			idx := m.selectedLLMIndex()
			if idx >= 0 {
				m.probing = true
				m.probeGen++
				gen := m.probeGen
				m.lastProbeIdx = idx
				cmds = append(cmds, m.probeCmd(idx, gen))
			}
			return m, tea.Batch(cmds...)
		}
		if key == "/" && m.tab == TabOverview {
			m.tab = TabLLM
			m.focusDetail = false
			m.syncViewports()
		}
		if m.tab == TabOverview && m.handleOverviewKey(key) {
			return m, nil
		}
		if m.handleTabKey(key) {
			m.syncViewports()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.resizeLists()
		m.syncViewports()
		return m, nil

	case snapshotMsg:
		if msg.gen != m.gen {
			return m, nil
		}
		m.loading = false
		m.snap = msg.snap
		m.history.record(msg.snap)
		m.overviewFocus = m.overviewFocus.clamp(msg.snap)
		m.syncLists()
		m.syncViewports()
		return m, nil

	case probeResultMsg:
		if msg.gen != m.probeGen {
			return m, nil
		}
		m.probing = false
		m.probeResult = &msg.result
		m.lastProbeIdx = msg.index
		if m.snap != nil && msg.index >= 0 && msg.index < len(m.snap.LLMs) {
			m.snap.LLMs[msg.index].Probe = &msg.result
		}
		m.syncViewports()
		return m, nil

	case tickMsg:
		if !m.loading {
			m.startRefresh()
			cmds = append(cmds, collectSnapshotCmd(m.ctx, m.opts.Timeout, m.gen, false))
		}
		cmds = append(cmds, tickCmd(m.opts.Refresh))
		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	if m.tabUsesList() && !m.focusDetail {
		var cmd tea.Cmd
		switch m.tab {
		case TabGPU:
			m.gpuList, cmd = m.gpuList.Update(msg)
		case TabLLM:
			m.llmList, cmd = m.llmList.Update(msg)
		case TabDocker:
			m.dockerList, cmd = m.dockerList.Update(msg)
		case TabDatabase:
			m.dbList, cmd = m.dbList.Update(msg)
		}
		cmds = append(cmds, cmd)
		m.syncDetailFromSelection()
	} else if m.tabUsesDetail() && m.tab != TabOverview {
		var cmd tea.Cmd
		switch m.tab {
		case TabGPU:
			m.gpuDetail, cmd = m.gpuDetail.Update(msg)
		case TabLLM:
			m.llmDetail, cmd = m.llmDetail.Update(msg)
		case TabDocker:
			m.dockerDetail, cmd = m.dockerDetail.Update(msg)
		case TabDatabase:
			m.dbDetail, cmd = m.dbDetail.Update(msg)
		case TabProbe:
			m.probeVP, cmd = m.probeVP.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "enter":
			if m.tabUsesList() {
				m.focusDetail = !m.focusDetail
			}
		case "esc":
			m.focusDetail = false
		case "left", "h":
			if m.tabUsesList() {
				m.focusDetail = false
			}
		case "right", "l":
			if m.tabUsesList() {
				m.focusDetail = true
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleOverviewKey(key string) bool {
	switch key {
	case "up", "k":
		m.overviewFocus = m.overviewFocus.moveUp(m.snap)
		return true
	case "down", "j":
		m.overviewFocus = m.overviewFocus.moveDown(m.snap)
		return true
	case "enter", "l", "right":
		if target, ok := m.overviewFocus.diveTarget(m.snap); ok {
			m.jumpToTarget(target)
		}
		return true
	case "2":
		m.tab = TabGPU
		m.focusDetail = false
		if m.overviewFocus.section == sectionGPU && m.snap != nil && len(m.snap.GPUs) > 0 {
			m.selectListAt(&m.gpuList, m.overviewFocus.row)
			m.syncDetailFromSelection()
		}
		return true
	case "3":
		m.tab = TabLLM
		m.focusDetail = false
		if m.overviewFocus.section == sectionLLM && m.snap != nil && len(m.snap.LLMs) > 0 {
			m.selectListAt(&m.llmList, m.overviewFocus.row)
			m.syncDetailFromSelection()
		}
		return true
	case "4":
		m.tab = TabDocker
		m.focusDetail = false
		if target, ok := m.overviewFocus.diveTarget(m.snap); ok && target.tab == TabDocker {
			m.jumpToTarget(target)
		}
		return true
	case "5":
		m.tab = TabDatabase
		m.focusDetail = false
		if m.overviewFocus.section == sectionDatabase && m.snap != nil && len(m.snap.Databases) > 0 {
			m.selectListAt(&m.dbList, m.overviewFocus.row)
			m.syncDetailFromSelection()
		}
		return true
	}
	return false
}

func (m *Model) jumpToTarget(target diveTarget) {
	m.tab = target.tab
	m.focusDetail = false
	switch target.tab {
	case TabGPU:
		m.selectListAt(&m.gpuList, target.index)
	case TabLLM:
		m.selectListAt(&m.llmList, target.index)
	case TabDocker:
		m.selectListAt(&m.dockerList, target.index)
	case TabDatabase:
		m.selectListAt(&m.dbList, target.index)
	}
	m.syncDetailFromSelection()
}

func (m *Model) selectListAt(l *list.Model, index int) {
	if index < 0 {
		return
	}
	for i, item := range l.Items() {
		if li, ok := item.(listItem); ok && li.index == index {
			l.Select(i)
			return
		}
	}
	if index < len(l.Items()) {
		l.Select(index)
	}
}

func (m *Model) handleTabKey(key string) bool {
	switch key {
	case "1":
		m.tab = TabOverview
		m.focusDetail = false
		return true
	case "2":
		m.tab = TabGPU
		m.focusDetail = false
		return true
	case "3":
		m.tab = TabLLM
		m.focusDetail = false
		return true
	case "4":
		m.tab = TabDocker
		m.focusDetail = false
		return true
	case "5":
		m.tab = TabDatabase
		m.focusDetail = false
		return true
	case "6":
		m.tab = TabProbe
		m.focusDetail = false
		return true
	case "tab":
		m.tab = Tab((int(m.tab) + 1) % tabCount)
		m.focusDetail = false
		return true
	case "shift+tab":
		n := int(m.tab) - 1
		if n < 0 {
			n = tabCount - 1
		}
		m.tab = Tab(n)
		m.focusDetail = false
		return true
	}
	return false
}

func (m *Model) startRefresh() {
	m.loading = true
	m.gen++
}

func (m *Model) tabUsesList() bool {
	switch m.tab {
	case TabGPU, TabLLM, TabDocker, TabDatabase:
		return true
	default:
		return false
	}
}

func (m *Model) tabUsesDetail() bool {
	return true
}

func (m *Model) contentHeight() int {
	h := m.height - headerLines - tabBarLines - footerLines
	if h < 1 {
		return 1
	}
	return h
}

func (m *Model) listWidth() int {
	w := m.width / 3
	if w < 24 {
		w = 24
	}
	if w > m.width-20 {
		w = m.width / 2
	}
	return w
}

func (m *Model) detailWidth() int {
	w := m.width - m.listWidth() - 1
	if w < 20 {
		return 20
	}
	return w
}

func (m *Model) resizeLists() {
	h := m.contentHeight()
	w := m.listWidth()
	m.gpuList.SetSize(w, h)
	m.llmList.SetSize(w, h)
	m.dockerList.SetSize(w, h)
	m.dbList.SetSize(w, h)
}

func (m *Model) syncViewports() {
	h := m.contentHeight()
	if m.tab == TabProbe {
		m.probeVP.Width = m.width - 2
		m.probeVP.Height = h
	} else if m.tabUsesList() {
		dw := m.detailWidth()
		m.gpuDetail.Width = dw
		m.gpuDetail.Height = h
		m.llmDetail.Width = dw
		m.llmDetail.Height = h
		m.dockerDetail.Width = dw
		m.dockerDetail.Height = h
		m.dbDetail.Width = dw
		m.dbDetail.Height = h
	}

	if m.snap != nil {
		m.probeVP.SetContent(probePanelText(m.snap, m.selectedLLMIndex(), m.probeResult, m.probing, m.styles))
	}
	m.syncDetailFromSelection()
}

func (m *Model) syncLists() {
	if m.snap == nil {
		return
	}

	barW := barWidthForTerminal(m.width)

	gpuItems := make([]list.Item, 0, len(m.snap.GPUs))
	for i, gpu := range m.snap.GPUs {
		gpuItems = append(gpuItems, listItem{
			title:  gpuListTitle(gpu, m.styles, barW),
			filter: gpuFilterValue(gpu),
			index:  i,
		})
	}
	m.gpuList.SetItems(gpuItems)
	clampListIndex(&m.gpuList, len(gpuItems))

	llmItems := make([]list.Item, 0, len(m.snap.LLMs))
	for i, llm := range m.snap.LLMs {
		llmItems = append(llmItems, listItem{
			title:  llmListTitle(llm, m.styles),
			filter: llmFilterValue(llm),
			index:  i,
		})
	}
	m.llmList.SetItems(llmItems)
	clampListIndex(&m.llmList, len(llmItems))

	var containers []snapshot.DockerContainer
	if m.snap.Docker != nil {
		containers = dockerContainersForList(m.snap.Docker.Containers, m.dockerShowAll)
	}
	dockerItems := make([]list.Item, 0, len(containers))
	for i, c := range containers {
		dockerItems = append(dockerItems, listItem{
			title:  dockerListTitle(c, m.styles),
			filter: dockerFilterValue(c),
			index:  i,
		})
	}
	m.dockerList.SetItems(dockerItems)
	clampListIndex(&m.dockerList, len(dockerItems))

	dbItems := make([]list.Item, 0, len(m.snap.Databases))
	for i, db := range m.snap.Databases {
		dbItems = append(dbItems, listItem{
			title:  dbListTitle(db, m.styles),
			filter: dbFilterValue(db),
			index:  i,
		})
	}
	m.dbList.SetItems(dbItems)
	clampListIndex(&m.dbList, len(dbItems))
}

func clampListIndex(l *list.Model, n int) {
	if n == 0 {
		l.Select(0)
		return
	}
	if l.Index() >= n {
		l.Select(n - 1)
	}
}

func (m *Model) syncDetailFromSelection() {
	if m.snap == nil {
		return
	}
	if idx, ok := m.selectedGPUIndex(); ok {
		m.gpuDetail.SetContent(gpuDetailText(m.snap.GPUs[idx], m.styles))
	}
	if idx, ok := m.selectedLLMIndexOK(); ok {
		m.llmDetail.SetContent(llmDetailText(m.snap.LLMs[idx], m.styles))
	}
	if idx, ok := m.selectedDockerIndex(); ok && m.snap.Docker != nil {
		containers := dockerContainersForList(m.snap.Docker.Containers, m.dockerShowAll)
		if idx < len(containers) {
			m.dockerDetail.SetContent(dockerDetailText(containers[idx], m.styles))
		}
	}
	if idx, ok := m.selectedDBIndex(); ok {
		m.dbDetail.SetContent(dbDetailText(m.snap.Databases[idx], m.styles))
	}
}

func (m *Model) selectedGPUIndex() (int, bool) {
	item, ok := m.gpuList.SelectedItem().(listItem)
	if !ok || m.snap == nil || item.index >= len(m.snap.GPUs) {
		return 0, false
	}
	return item.index, true
}

func (m *Model) selectedLLMIndex() int {
	idx, ok := m.selectedLLMIndexOK()
	if !ok {
		return -1
	}
	return idx
}

func (m *Model) selectedLLMIndexOK() (int, bool) {
	item, ok := m.llmList.SelectedItem().(listItem)
	if !ok || m.snap == nil || item.index >= len(m.snap.LLMs) {
		if m.snap != nil && len(m.snap.LLMs) > 0 {
			return 0, true
		}
		return 0, false
	}
	return item.index, true
}

func (m *Model) selectedDockerIndex() (int, bool) {
	item, ok := m.dockerList.SelectedItem().(listItem)
	if !ok || m.snap == nil || m.snap.Docker == nil {
		return 0, false
	}
	containers := dockerContainersForList(m.snap.Docker.Containers, m.dockerShowAll)
	if item.index >= len(containers) {
		return 0, false
	}
	return item.index, true
}

func (m *Model) selectedDBIndex() (int, bool) {
	item, ok := m.dbList.SelectedItem().(listItem)
	if !ok || m.snap == nil || item.index >= len(m.snap.Databases) {
		if m.snap != nil && len(m.snap.Databases) > 0 {
			return 0, true
		}
		return 0, false
	}
	return item.index, true
}

func (m *Model) probeCmd(index int, gen uint64) tea.Cmd {
	var llm snapshot.LLM
	if m.snap != nil && index >= 0 && index < len(m.snap.LLMs) {
		llm = m.snap.LLMs[index]
	}
	timeout := m.opts.Timeout
	parent := m.ctx
	return func() tea.Msg {
		if llm.Endpoint == "" {
			return probeResultMsg{gen: gen, index: index, result: snapshot.ProbeResult{OK: false, Error: "no endpoint"}}
		}
		ctx, cancel := context.WithTimeout(parent, timeout)
		defer cancel()
		client := probe.NewClient(ctx)
		result := probe.SmokeLLM(ctx, client, &llm, "", "Hello")
		return probeResultMsg{gen: gen, index: index, result: result}
	}
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderTabBar())
	b.WriteString("\n")
	b.WriteString(m.renderBody())
	b.WriteString("\n")
	b.WriteString(m.renderFooter())
	if m.help {
		b.WriteString("\n")
		b.WriteString(m.renderHelp())
	}
	return b.String()
}

func (m Model) renderHelp() string {
	switch m.tab {
	case TabOverview:
		return m.styles.muted.Render("Overview: ↑/↓ sections · enter dive · / search LLMs · 2-6 tabs · r refresh · q quit")
	default:
		return m.styles.muted.Render("1-6 tabs · ↑/↓ j/k select · / filter · enter detail · a all containers (docker) · p probe · r refresh · q quit")
	}
}

func (m Model) renderHeader() string {
	title := m.styles.title.Render("bareai")
	host := "unknown"
	collected := "-"
	if m.snap != nil {
		if m.snap.Host != nil && m.snap.Host.Hostname != "" {
			host = m.snap.Host.Hostname
		}
		collected = m.snap.CollectedAt.Format("15:04:05")
	}
	sub := m.styles.subtitle.Render(fmt.Sprintf(" %s · collected %s", host, collected))
	status := ""
	if m.loading {
		status = " " + m.spinner.View() + " refreshing"
	}
	line1 := title + sub + m.styles.header.Render(status)
	return line1
}

func (m Model) renderTabBar() string {
	labels := []string{"1 Overview", "2 GPUs", "3 LLMs", "4 Docker", "5 DBs", "6 Probe"}
	parts := make([]string, len(labels))
	for i, label := range labels {
		if Tab(i) == m.tab {
			parts[i] = m.styles.tabActive.Render(label)
		} else {
			parts[i] = m.styles.tab.Render(label)
		}
	}
	return strings.Join(parts, " ")
}

func (m Model) renderBody() string {
	h := m.contentHeight()
	if m.snap == nil && m.loading {
		return m.styles.pane.Render(strings.Repeat("\n", h-1) + "Collecting infrastructure snapshot...")
	}

	switch m.tab {
	case TabOverview:
		return m.styles.pane.Render(renderDashboard(m.snap, m.history, m.overviewFocus, m.width, m.styles))
	case TabProbe:
		return m.styles.pane.Render(m.probeVP.View())
	case TabGPU:
		if len(m.snap.GPUs) == 0 {
			return m.styles.pane.Render(render.EmptyHint("gpu"))
		}
		return m.renderSplit(m.gpuList.View(), m.gpuDetail.View())
	case TabLLM:
		if len(m.snap.LLMs) == 0 {
			return m.styles.pane.Render(render.EmptyHint("llm"))
		}
		return m.renderSplit(m.llmList.View(), m.llmDetail.View())
	case TabDocker:
		if m.snap != nil && (m.snap.Docker == nil || !m.snap.Docker.Available) {
			return m.styles.pane.Render(render.EmptyHint("docker"))
		}
		if m.snap != nil && len(dockerContainersForList(m.snap.Docker.Containers, m.dockerShowAll)) == 0 {
			return m.renderSplit(m.dockerList.View(), "No running containers.")
		}
		return m.renderSplit(m.dockerList.View(), m.dockerDetail.View())
	case TabDatabase:
		if len(m.snap.Databases) == 0 {
			return m.styles.pane.Render(render.EmptyHint("db"))
		}
		return m.renderSplit(m.dbList.View(), m.dbDetail.View())
	default:
		return ""
	}
}

func (m Model) renderSplit(left, right string) string {
	if m.focusDetail {
		right = m.styles.border.Render(right)
	} else {
		left = m.styles.border.Render(left)
	}
	return lipglossJoinHorizontal(m.listWidth(), m.detailWidth(), left, right)
}

func (m Model) renderFooter() string {
	var hint string
	switch m.tab {
	case TabOverview:
		hint = "↑/↓ sections · enter dive · / search · r refresh · q quit · ? help"
	case TabDocker:
		if m.tabUsesList() && m.focusDetail {
			hint = "detail · esc back · a all/running · / filter · r refresh · q quit"
		} else {
			hint = "↑/↓ select · / filter · a all/running · enter detail · r refresh · q quit · ? help"
		}
	default:
		if m.tabUsesList() && m.focusDetail {
			hint = "detail focused · ↑/↓ scroll · esc back · r refresh · q quit"
		} else {
			hint = "↑/↓ select · / filter · enter detail · p probe · r refresh · q quit · ? help"
		}
	}
	return m.styles.footer.Render(hint)
}

func lipglossJoinHorizontal(leftW, rightW int, left, right string) string {
	// Avoid importing lipgloss in every file; simple side-by-side join.
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	h := len(leftLines)
	if len(rightLines) > h {
		h = len(rightLines)
	}
	out := make([]string, 0, h)
	for i := 0; i < h; i++ {
		l := ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		r := ""
		if i < len(rightLines) {
			r = rightLines[i]
		}
		pad := leftW - len(l)
		if pad < 1 {
			pad = 1
		}
		out = append(out, l+strings.Repeat(" ", pad)+r)
	}
	return strings.Join(out, "\n")
}
