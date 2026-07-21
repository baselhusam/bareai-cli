package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/baselhusam/bareai-cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration helpers",
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the resolved config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.Path()
		if path == "" {
			return fmt.Errorf("could not determine config path")
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), path)
		return err
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
}
