package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const (
	secretSavedMsg   = "secret saved"
	secretDeletedMsg = "secret deleted"
)

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [service] [user]",
		Short: "Set secret in keyring",
		Long: longDesc(`
			Sets a secret for a service/user combination in the keyring.

			If stdin is a pipe, the secret is read from there. Otherwise it will prompt for the secret interactively.`),
		Example: example(`
			# Secret via stdin
			$ echo -n "supersecret" | keyring set myservice myuser

			# Secret via interactive prompt
			$ keyring set myservice myuser
			Enter Secret:`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			fi, err := os.Stdin.Stat()
			if err != nil {
				return err
			}

			var secret []byte

			if (fi.Mode() & os.ModeCharDevice) == 0 {
				secret, err = ioutil.ReadAll(os.Stdin)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), "Enter Secret: ")

				secret, err = term.ReadPassword(int(os.Stdin.Fd()))

				fmt.Fprintln(cmd.OutOrStdout())
			}

			if err != nil {
				return err
			}

			err = keyring.Set(service, user, string(secret))
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), secretSavedMsg)

			return nil
		},
	}

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [service] [user]",
		Short: "Read secret from keyring",
		Long: longDesc(`
			Reads a secret for a service/user combination from the keyring and writes it to stdout.

			The returned secret is terminated by a newline character.`),
		Example: example(`
			$ keyring get myservice myuser`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			secret, err := keyring.Get(service, user)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), secret)

			return nil
		},
	}

	return cmd
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [service] [user]",
		Short: "Delete secret from keyring",
		Long: longDesc(`
			Deletes the secret for a service/user combination from the keyring.`),
		Example: example(`
			$ keyring delete myservice myuser`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			err := keyring.Delete(service, user)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), secretDeletedMsg)

			return nil
		},
	}

	return cmd
}
