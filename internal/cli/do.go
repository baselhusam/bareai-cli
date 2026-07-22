package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/baselhusam/bareai-cli/internal/action"
	"github.com/baselhusam/bareai-cli/internal/config"
	"github.com/baselhusam/bareai-cli/internal/render"
)

var doOpts struct {
	Finding     string
	Container   string
	Endpoint    string
	Tail        int
	Yes         bool
	DryRun      bool
	NoReprobe   bool
}

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Confirm-gated actions tied to doctor findings",
	Long: `Run a small set of finding-scoped actions after diagnosis.

Mutating verbs require --finding and confirmation (--yes or interactive prompt).
Use "bareai do plan <verb>" or --dry-run to preview changes without side effects.

MCP remains read-only; actions are CLI-only in Phase 12.`,
	Example: `  bareai do list
  bareai do plan restart --finding llm.unreachable --container ollama
  bareai do restart --finding llm.unreachable --container ollama --yes
  bareai do logs --finding llm.unreachable --tail 200
  bareai do reprobe --finding llm.no_models --endpoint http://127.0.0.1:11434
  bareai do free-gpu --finding gpu.vram_high --container vllm --yes`,
}

var doListCmd = &cobra.Command{
	Use:   "list",
	Short: "List actions available for current findings",
	RunE:  runDoList,
}

var doPlanCmd = &cobra.Command{
	Use:   "plan [verb]",
	Short: "Preview an action without side effects",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDo(cmd, args[0], true)
	},
}

func init() {
	def := config.Default()
	doOpts.Tail = def.Actions.LogTail
	doCmd.AddCommand(doListCmd, doPlanCmd)
	for _, verb := range []string{action.VerbLogs, action.VerbReprobe, action.VerbRestart, action.VerbStop, action.VerbFreeGPU} {
		doCmd.AddCommand(newDoVerbCmd(verb))
	}
	doCmd.PersistentFlags().StringVar(&doOpts.Finding, "finding", "", "doctor finding ID (required for mutate verbs)")
	doCmd.PersistentFlags().StringVar(&doOpts.Container, "container", "", "container name or ID from snapshot")
	doCmd.PersistentFlags().StringVar(&doOpts.Endpoint, "endpoint", "", "LLM endpoint for reprobe")
	doCmd.PersistentFlags().IntVar(&doOpts.Tail, "tail", def.Actions.LogTail, "log lines for logs verb")
	doCmd.PersistentFlags().BoolVar(&doOpts.Yes, "yes", false, "skip interactive confirmation for mutating verbs")
	doCmd.PersistentFlags().BoolVar(&doOpts.DryRun, "dry-run", false, "preview only; no side effects")
	doCmd.PersistentFlags().BoolVar(&doOpts.NoReprobe, "no-reprobe", false, "skip automatic reprobe after restart/free-gpu")
}

func newDoVerbCmd(verb string) *cobra.Command {
	return &cobra.Command{
		Use:   verb,
		Short: fmt.Sprintf("Run %s action", verb),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDo(cmd, verb, false)
		},
	}
}

func runDoList(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
	defer cancel()
	snap := action.CollectSnapshotForList(ctx)
	entries := action.ListAvailable(snap)
	if opts.JSON {
		return render.WriteActionListJSON(cmd.OutOrStdout(), entries)
	}
	return render.WriteActionList(cmd.OutOrStdout(), entries)
}

func runDo(cmd *cobra.Command, verb string, plan bool) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), opts.Timeout)
	defer cancel()

	cfg := config.Global()
	req := action.Request{
		Verb:        verb,
		FindingID:   doOpts.Finding,
		Container:   doOpts.Container,
		Endpoint:    doOpts.Endpoint,
		Tail:        doOpts.Tail,
		DryRun:      doOpts.DryRun,
		PlanOnly:    plan,
		AutoReprobe: cfg.Actions.AutoReprobe && !doOpts.NoReprobe,
		LogMaxBytes: cfg.Actions.LogMaxBytes,
	}

	if action.Mutates(verb) && !plan && !doOpts.DryRun {
		preview, err := (&action.Executor{}).Run(ctx, action.Request{
			Verb:        verb,
			FindingID:   req.FindingID,
			Container:   req.Container,
			Endpoint:    req.Endpoint,
			Tail:        req.Tail,
			PlanOnly:    true,
			AutoReprobe: req.AutoReprobe,
			LogMaxBytes: req.LogMaxBytes,
		})
		if err != nil {
			return err
		}
		if !opts.JSON {
			if err := render.WriteAction(cmd.OutOrStdout(), preview); err != nil {
				return err
			}
		}
		if cfg.Actions.Confirm && !doOpts.Yes {
			if !term.IsTerminal(int(os.Stdin.Fd())) {
				return fmt.Errorf("confirmation required: use --yes in non-interactive mode")
			}
			if _, err := fmt.Fprint(cmd.OutOrStdout(), "Proceed? [y/N]: "); err != nil {
				return err
			}
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			answer := strings.TrimSpace(strings.ToLower(line))
			if answer != "y" && answer != "yes" {
				return fmt.Errorf("cancelled")
			}
		}
		req.Confirmed = true
	}

	res, err := (&action.Executor{}).Run(ctx, req)
	if opts.JSON {
		if writeErr := render.WriteJSON(cmd.OutOrStdout(), res); writeErr != nil {
			return writeErr
		}
	} else if err := render.WriteAction(cmd.OutOrStdout(), res); err != nil {
		return err
	}
	if err != nil {
		return err
	}
	if !res.OK {
		return fmt.Errorf("action failed")
	}
	return nil
}
