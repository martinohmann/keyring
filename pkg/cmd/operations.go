package cmd

import (
	"fmt"
	"io"
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

			secret, err := readSecret(cmd.InOrStdin())
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

			return writeSecret(cmd.OutOrStdout(), secret)
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

type fder interface {
	Fd() uintptr
}

// writeSecret writes the secret to w. If w is a terminal, the secret will be
// newline-terminated.
func writeSecret(w io.Writer, secret string) (err error) {
	f, ok := w.(fder)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		_, err = w.Write([]byte(secret))
	} else {
		_, err = fmt.Fprintln(w, secret)
	}

	return
}

// readSecret reads the secret from r. If r is a terminal, the user will be
// prompted to enter it interactively.
func readSecret(r io.Reader) ([]byte, error) {
	f, ok := r.(fder)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return ioutil.ReadAll(r)
	}

	fmt.Fprint(os.Stdout, "Enter Secret: ")
	defer fmt.Fprintln(os.Stdout)

	return term.ReadPassword(int(f.Fd()))
}
