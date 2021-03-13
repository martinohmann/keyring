package cmd

import (
	"fmt"
	"os"

	"github.com/martinohmann/exit"
	"github.com/spf13/cobra"
)

// Version is set via build args.
var version = "v0.0.0-master"

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "keyring",
		Short:         "Interact with the operating system's keyring.",
		Version:       version,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Do not display usage on errors that happen after successful
			// command parsing.
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exit.Exit(err)
	}
}
