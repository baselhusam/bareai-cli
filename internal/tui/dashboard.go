package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

type overviewSection int

const (
	sectionHost overviewSection = iota
	sectionGPU
	sectionLLM
	sectionCorrelation
	sectionFindings
	sectionSkipped
	overviewSectionCount
)

type overviewFocus struct {
	section overviewSection
	row     int
}

func (f overviewFocus) clamp(snap *snapshot.Snapshot) overviewFocus {
	if snap == nil {
		return overviewFocus{section: sectionHost, row: 0}
	}
	maxSection := sectionSkipped
	if len(snap.Skipped) == 0 {
		maxSection = sectionFindings
	}
	if len(snap.Findings) == 0 {
		maxSection = sectionCorrelation
	}
	if f.section > maxSection {
		f.section = maxSection
	}
	rows := sectionRowCount(snap, f.section)
	if rows == 0 {
		f.row = 0
	} else if f.row >= rows {
		f.row = rows - 1
	}
	if f.row < 0 {
		f.row = 0
	}
	return f
}

func sectionRowCount(snap *snapshot.Snapshot, sec overviewSection) int {
	if snap == nil {
		return 1
	}
	switch sec {
	case sectionHost:
		return 1
	case sectionGPU:
		if len(snap.GPUs) == 0 {
			return 1
		}
		return len(snap.GPUs)
	case sectionLLM:
		if len(snap.LLMs) == 0 {
			return 1
		}
		return len(snap.LLMs)
	case sectionCorrelation:
		if len(snap.Correlations) == 0 {
			return 1
		}
		return len(snap.Correlations)
	case sectionFindings:
		limit := 3
		if len(snap.Findings) < limit {
			return len(snap.Findings)
		}
		return limit
	case sectionSkipped:
		return len(snap.Skipped)
	default:
		return 1
	}
}

func (f overviewFocus) moveUp(snap *snapshot.Snapshot) overviewFocus {
	f = f.clamp(snap)
	if f.row > 0 {
		f.row--
		return f
	}
	if f.section > sectionHost {
		f.section--
		f.row = sectionRowCount(snap, f.section) - 1
		if f.row < 0 {
			f.row = 0
		}
	}
	return f.clamp(snap)
}

func (f overviewFocus) moveDown(snap *snapshot.Snapshot) overviewFocus {
	f = f.clamp(snap)
	rows := sectionRowCount(snap, f.section)
	if f.row < rows-1 {
		f.row++
		return f
	}
	maxSection := sectionSkipped
	if snap != nil && len(snap.Skipped) == 0 {
		maxSection = sectionFindings
	}
	if snap != nil && len(snap.Findings) == 0 {
		maxSection = sectionCorrelation
	}
	if f.section < maxSection {
		f.section++
		f.row = 0
	}
	return f.clamp(snap)
}

type diveTarget struct {
	tab   Tab
	index int
}

func (f overviewFocus) diveTarget(snap *snapshot.Snapshot) (diveTarget, bool) {
	if snap == nil {
		return diveTarget{}, false
	}
	switch f.section {
	case sectionGPU:
		if len(snap.GPUs) == 0 {
			return diveTarget{tab: TabGPU}, true
		}
		if f.row < len(snap.GPUs) {
			return diveTarget{tab: TabGPU, index: f.row}, true
		}
	case sectionLLM:
		if len(snap.LLMs) == 0 {
			return diveTarget{tab: TabLLM}, true
		}
		if f.row < len(snap.LLMs) {
			return diveTarget{tab: TabLLM, index: f.row}, true
		}
	case sectionCorrelation:
		if len(snap.Correlations) == 0 {
			return diveTarget{tab: TabLLM}, true
		}
		if f.row < len(snap.Correlations) {
			row := snap.Correlations[f.row]
			if row.ContainerName != "" && snap.Docker != nil {
				for i, c := range dockerContainersForList(snap.Docker.Containers, false) {
					if c.Name == row.ContainerName || strings.Contains(c.Name, row.ContainerName) {
						return diveTarget{tab: TabDocker, index: i}, true
					}
				}
			}
			for i, llm := range snap.LLMs {
				if llm.Endpoint == row.Endpoint {
					return diveTarget{tab: TabLLM, index: i}, true
				}
			}
			return diveTarget{tab: TabLLM}, true
		}
	}
	return diveTarget{}, false
}

func renderDashboard(snap *snapshot.Snapshot, hist metricHistory, focus overviewFocus, width int, s styles) string {
	if snap == nil {
		return s.muted.Render("Collecting infrastructure snapshot...")
	}
	focus = focus.clamp(snap)

	barW := barWidthForTerminal(width)
	sparkW := sparkWidthForTerminal(width)

	var parts []string
	parts = append(parts, renderHostPanel(snap, hist, focus, barW, sparkW, s))
	parts = append(parts, renderGPUPanel(snap, hist, focus, barW, sparkW, s))
	parts = append(parts, renderLLMPanel(snap, focus, width, s))
	parts = append(parts, renderCorrelationPanel(snap, focus, width, s))
	if len(snap.Findings) > 0 {
		parts = append(parts, renderFindingsPanel(snap, focus, s))
	}
	if len(snap.Skipped) > 0 {
		parts = append(parts, renderSkippedPanel(snap, focus, s))
	}
	return strings.Join(parts, "\n")
}

func panelTitle(s styles, title string, sec overviewSection, focus overviewFocus) string {
	label := s.label.Render(title)
	if focus.section == sec {
		return s.focus.Render("▸ ") + label
	}
	return "  " + label
}

func rowPrefix(focus overviewFocus, sec overviewSection, row int) string {
	if focus.section == sec && focus.row == row {
		return "▸ "
	}
	return "  "
}

func renderHostPanel(snap *snapshot.Snapshot, hist metricHistory, focus overviewFocus, barW, sparkW int, s styles) string {
	title := panelTitle(s, "Host", sectionHost, focus)
	var lines []string
	lines = append(lines, title)

	if snap.Host == nil {
		lines = append(lines, s.muted.Render("  unavailable"))
		return wrapPanel(lines, focus.section == sectionHost, s)
	}

	h := snap.Host
	load := loadPct(h.Load1, h.CPUCores)
	mem := pctUsed(h.MemUsed, h.MemTotal)
	cpuLine := fmt.Sprintf("%s  load %.2f", truncate(h.CPUModel, 28), h.Load1)
	if h.CPUCores > 0 {
		cpuLine = fmt.Sprintf("%s (%d cores)  load %.2f", truncate(h.CPUModel, 22), h.CPUCores, h.Load1)
	}
	lines = append(lines, fmt.Sprintf("%s%s  %s %s",
		rowPrefix(focus, sectionHost, 0),
		cpuLine,
		renderBar(s, load, barW),
		renderSparkline(s, hist.hostLoad(), sparkW),
	))
	lines = append(lines, fmt.Sprintf("%s  RAM %s/%s  %s %s",
		rowPrefix(focus, sectionHost, 0),
		formatBytes(h.MemUsed), formatBytes(h.MemTotal),
		renderBar(s, mem, barW),
		renderSparkline(s, hist.hostMem(), sparkW),
	))
	if disk := primaryDisk(h); disk != nil {
		diskPct := pctUsed(disk.Used, disk.Total)
		lines = append(lines, fmt.Sprintf("%s  Disk %s  %s %s",
			rowPrefix(focus, sectionHost, 0),
			disk.Mount,
			renderBar(s, diskPct, barW),
			s.value.Render(fmt.Sprintf("%.0f%%", diskPct)),
		))
	}
	return wrapPanel(lines, focus.section == sectionHost, s)
}

func primaryDisk(h *snapshot.Host) *snapshot.Disk {
	if h == nil || len(h.Disks) == 0 {
		return nil
	}
	for _, d := range h.Disks {
		if d.Mount == "/" {
			return &d
		}
	}
	return &h.Disks[0]
}

func renderGPUPanel(snap *snapshot.Snapshot, hist metricHistory, focus overviewFocus, barW, sparkW int, s styles) string {
	title := panelTitle(s, "GPUs", sectionGPU, focus)
	var lines []string
	lines = append(lines, title)

	if len(snap.GPUs) == 0 {
		lines = append(lines, s.muted.Render("  no accelerators detected"))
		return wrapPanel(lines, focus.section == sectionGPU, s)
	}

	for i, gpu := range snap.GPUs {
		util := 0.0
		utilLabel := "n/a"
		if gpu.Utilization != nil {
			util = *gpu.Utilization
			utilLabel = fmt.Sprintf("%.0f%%", util)
		}
		memPct := pctUsed(gpu.MemoryUsed, gpu.MemoryTotal)
		memLabel := "unified"
		if gpu.MemoryTotal > 0 {
			memLabel = fmt.Sprintf("%s/%s", formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal))
		}
		temp := ""
		if gpu.Temperature != nil {
			temp = tempStyle(s, *gpu.Temperature).Render(fmt.Sprintf("%.0f°C", *gpu.Temperature))
		}
		line := fmt.Sprintf("%s[%d] %s  util %s %s %s  vram %s %s %s",
			rowPrefix(focus, sectionGPU, i),
			gpu.Index,
			truncate(gpu.Name, 18),
			utilLabel,
			renderBar(s, util, barW),
			renderSparkline(s, hist.gpuUtil(gpu.Index), sparkW),
			memLabel,
			renderBar(s, memPct, barW),
			renderSparkline(s, hist.gpuMem(gpu.Index), sparkW),
		)
		if temp != "" {
			line += "  " + temp
		}
		lines = append(lines, line)
	}
	return wrapPanel(lines, focus.section == sectionGPU, s)
}

func renderLLMPanel(snap *snapshot.Snapshot, focus overviewFocus, width int, s styles) string {
	title := panelTitle(s, "Providers / LLMs", sectionLLM, focus)
	var lines []string
	lines = append(lines, title)

	if len(snap.LLMs) == 0 {
		lines = append(lines, s.muted.Render("  none discovered"))
		return wrapPanel(lines, focus.section == sectionLLM, s)
	}

	epW := 22
	if width >= 100 {
		epW = 28
	}
	for i, llm := range snap.LLMs {
		pid := "-"
		if llm.PID > 0 {
			pid = fmt.Sprintf("%d", llm.PID)
		}
		gpu := "-"
		if llm.GPUIndex != nil {
			gpu = fmt.Sprintf("%d", *llm.GPUIndex)
		}
		health := s.muted.Render("?")
		if llm.Health != nil {
			label := "fail"
			if llm.Health.OK {
				label = "ok"
			}
			health = s.healthStyle(llm.Health.OK).Render(label)
		}
		name := llm.Name
		if name == "" {
			name = llm.Runtime
		}
		line := fmt.Sprintf("%s%-10s %-14s %-*s pid %-6s gpu %-3s %s",
			rowPrefix(focus, sectionLLM, i),
			truncate(llm.Runtime, 10),
			truncate(name, 14),
			epW, truncate(llm.Endpoint, epW),
			pid, gpu, health,
		)
		lines = append(lines, line)
	}
	return wrapPanel(lines, focus.section == sectionLLM, s)
}

func renderCorrelationPanel(snap *snapshot.Snapshot, focus overviewFocus, width int, s styles) string {
	title := panelTitle(s, "Correlation", sectionCorrelation, focus)
	var lines []string
	lines = append(lines, title)

	if len(snap.Correlations) == 0 {
		lines = append(lines, s.muted.Render("  none"))
		return wrapPanel(lines, focus.section == sectionCorrelation, s)
	}

	epW := 20
	if width >= 100 {
		epW = 24
	}
	for i, row := range snap.Correlations {
		pid := "-"
		if row.PID > 0 {
			pid = fmt.Sprintf("%d", row.PID)
		}
		gpu := "-"
		if row.GPUIndex != nil {
			gpu = fmt.Sprintf("%d", *row.GPUIndex)
		}
		container := row.ContainerName
		if container == "" {
			container = "-"
		}
		vram := "-"
		if row.VRAMBytes > 0 {
			vram = formatBytes(row.VRAMBytes)
		}
		line := fmt.Sprintf("%s%-*s %-8s %-10s pid %-5s gpu %-3s vram %s",
			rowPrefix(focus, sectionCorrelation, i),
			epW, truncate(row.Endpoint, epW),
			truncate(row.Runtime, 8),
			truncate(container, 10),
			pid, gpu, vram,
		)
		lines = append(lines, line)
	}
	return wrapPanel(lines, focus.section == sectionCorrelation, s)
}

func renderFindingsPanel(snap *snapshot.Snapshot, focus overviewFocus, s styles) string {
	title := panelTitle(s, "Findings", sectionFindings, focus)
	var lines []string
	lines = append(lines, title)

	limit := 3
	if len(snap.Findings) < limit {
		limit = len(snap.Findings)
	}
	for i := 0; i < limit; i++ {
		f := snap.Findings[i]
		sev := f.Severity
		if sev == "" {
			sev = "info"
		}
		line := fmt.Sprintf("%s[%s] %s: %s",
			rowPrefix(focus, sectionFindings, i),
			s.severityStyle(sev).Render(sev),
			f.ID,
			truncate(f.Summary, 60),
		)
		lines = append(lines, line)
	}
	if len(snap.Findings) > limit {
		lines = append(lines, s.muted.Render(fmt.Sprintf("  … %d more (bareai doctor)", len(snap.Findings)-limit)))
	}
	return wrapPanel(lines, focus.section == sectionFindings, s)
}

func renderSkippedPanel(snap *snapshot.Snapshot, focus overviewFocus, s styles) string {
	title := panelTitle(s, "Skipped", sectionSkipped, focus)
	var lines []string
	lines = append(lines, title)
	for i, skip := range snap.Skipped {
		lines = append(lines, fmt.Sprintf("%s%s: %s",
			rowPrefix(focus, sectionSkipped, i),
			s.muted.Render(skip.Component),
			truncate(skip.Reason, 50),
		))
	}
	return wrapPanel(lines, focus.section == sectionSkipped, s)
}

func wrapPanel(lines []string, focused bool, s styles) string {
	body := strings.Join(lines, "\n")
	if focused {
		return s.border.BorderForeground(lipgloss.Color("86")).Render(body)
	}
	return s.border.Render(body)
}
