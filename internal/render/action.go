package render

import (
	"fmt"
	"io"

	"github.com/baselhusam/bareai-cli/internal/action"
	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// WriteAction renders a human-readable action result.
func WriteAction(w io.Writer, res action.Result) error {
	if _, err := fmt.Fprintf(w, "bareai do %s", res.Verb); err != nil {
		return err
	}
	if res.FindingID != "" {
		if _, err := fmt.Fprintf(w, " (finding %s)", res.FindingID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if res.DryRun {
		if _, err := fmt.Fprintln(w, "Mode: plan/dry-run"); err != nil {
			return err
		}
	}
	if res.Target != nil {
		if _, err := fmt.Fprintf(w, "Target: %s %q\n", res.Target.Kind, targetLabel(res.Target)); err != nil {
			return err
		}
	}
	if res.Would != "" {
		if _, err := fmt.Fprintf(w, "Would: %s\n", res.Would); err != nil {
			return err
		}
	}
	if res.Before != nil {
		if _, err := fmt.Fprintf(w, "Before: %s (%s)\n", res.Before.Name, res.Before.State); err != nil {
			return err
		}
	}
	if res.After != nil {
		if _, err := fmt.Fprintf(w, "After: %s (%s)\n", res.After.Name, res.After.State); err != nil {
			return err
		}
	}
	for _, step := range res.Steps {
		line := fmt.Sprintf("Step %s: ok=%v", step.Verb, step.OK)
		if step.Summary != "" {
			line += " — " + step.Summary
		}
		if step.Error != "" {
			line += " (" + step.Error + ")"
		}
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	if res.Output != "" {
		if _, err := fmt.Fprintln(w, "\nLogs:"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, res.Output); err != nil {
			return err
		}
	}
	status := "ok"
	if !res.OK {
		status = "failed"
	}
	if _, err := fmt.Fprintf(w, "\nResult: %s", status); err != nil {
		return err
	}
	if res.Error != "" {
		if _, err := fmt.Fprintf(w, " (%s)", res.Error); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

// WriteActionList renders available actions.
func WriteActionList(w io.Writer, entries []action.ListEntry) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, "No confirm-gated actions available for current findings.")
		return err
	}
	for _, e := range entries {
		if _, err := fmt.Fprintf(w, "[%s] %s → %s (%s)\n", e.FindingID, e.Verb, e.TargetRef, e.Summary); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "  %s\n", e.Command); err != nil {
			return err
		}
	}
	return nil
}

// WriteActionListJSON writes list entries as JSON.
func WriteActionListJSON(w io.Writer, entries []action.ListEntry) error {
	return WriteJSON(w, map[string]any{"actions": entries})
}

func targetLabel(t *action.Target) string {
	if t.Name != "" {
		return t.Name
	}
	if t.Endpoint != "" {
		return t.Endpoint
	}
	return t.ID
}

// FormatFindingDo renders structured Do hints for doctor output.
func FormatFindingDo(f snapshot.Finding, width int) []string {
	if len(f.Do) == 0 {
		return nil
	}
	var parts []string
	for _, offer := range f.Do {
		parts = append(parts, action.PlanCommand(offer.Verb, f.ID, offer.TargetKind, offer.TargetRef))
	}
	text := "Do: " + joinParts(parts)
	return wrapText(text, width, "  ")
}

func joinParts(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += " · "
		}
		out += p
	}
	return out
}
