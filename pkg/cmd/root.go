package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "keyring",
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Do not display usage after successful command parsing.
			cmd.SilenceUsage = true
		},
	}

	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newDeleteCommand())

	return cmd
}

func Execute() {
	cmd := newRootCommand()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
