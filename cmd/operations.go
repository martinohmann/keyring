package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/martinohmann/exit"
	"github.com/spf13/cobra"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const (
	secretCreatedMsg = "secret created"
	secretUpdatedMsg = "secret updated"
	secretDeletedMsg = "secret deleted"
)

var (
	errAborted              = errors.New("operation aborted")
	errConfirmationRequired = errors.New("confirmation required")
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
			enter secret:`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			secret, err := readSecret(cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				return exit.Error(exit.CodeIOErr, err)
			}

			var msg string

			_, err = keyring.Get(service, user)
			switch {
			case errors.Is(err, keyring.ErrNotFound):
				msg = secretCreatedMsg
			case err == nil:
				msg = secretUpdatedMsg

				if err := confirm(cmd, "secret exists, overwrite?"); err != nil {
					return err
				}
			default:
				return err
			}

			err = keyring.Set(service, user, string(secret))
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), msg)

			return nil
		},
	}

	cmd.Flags().Bool("yes", false, "automatically confirm secret overwrite prompts")

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [service] [user]",
		Short: "Read secret from keyring",
		Long: longDesc(`
			Reads a secret for a service/user combination from the keyring and writes it to stdout.

			If stdout is a terminal a newline character is printed after the secret.`),
		Example: example(`
			# Write secret to stdout
			$ keyring get myservice myuser

			# Pipe secret into another command
			$ keyring get myservice myuser | cat`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			secret, err := keyring.Get(service, user)
			if err != nil {
				return err
			}

			err = writeSecret(cmd.OutOrStdout(), secret)
			if err != nil {
				return exit.Error(exit.CodeIOErr, err)
			}

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

			_, err := keyring.Get(service, user)
			if err != nil {
				return err
			}

			if err := confirm(cmd, "delete secret?"); err != nil {
				return err
			}

			err = keyring.Delete(service, user)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), secretDeletedMsg)

			return nil
		},
	}

	cmd.Flags().Bool("yes", false, "automatically confirm secret deletion prompts")

	return cmd
}

type fder interface {
	Fd() uintptr
}

// writeSecret writes the secret to out. If out is a terminal, the secret will
// be newline-terminated.
func writeSecret(out io.Writer, secret string) (err error) {
	f, ok := out.(fder)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		_, err = out.Write([]byte(secret))
	} else {
		_, err = fmt.Fprintln(out, secret)
	}

	return
}

// readSecret reads the secret from in. If in is a terminal, the user will be
// prompted to enter it interactively. The prompt is written to out in this
// case.
func readSecret(in io.Reader, out io.Writer) ([]byte, error) {
	fin, ok := in.(fder)
	if !ok || !term.IsTerminal(int(fin.Fd())) {
		return ioutil.ReadAll(in)
	}

	fmt.Fprint(out, "enter secret: ")
	defer fmt.Fprintln(out)

	return term.ReadPassword(int(fin.Fd()))
}

// ask prints question to out and waits for confirmation input on in. Returns
// errAborted if the user chose to stop or reading from in failed.
func ask(in io.Reader, out io.Writer, question string) error {
	fin, ok := in.(fder)
	if !ok || !term.IsTerminal(int(fin.Fd())) {
		return exit.Error(exit.CodeUsage, errConfirmationRequired)
	}

	state, err := term.MakeRaw(int(fin.Fd()))
	if err != nil {
		return exit.Errorf(exit.CodeIOErr, "failed to put terminal into raw mode: %w", err)
	}

	fmt.Fprintf(out, "%s [y/N] ", question)
	defer fmt.Fprintln(out)
	defer term.Restore(int(fin.Fd()), state) // nolint: errcheck

	r := bufio.NewReader(in)

	c, _, _ := r.ReadRune()
	switch c {
	case 'y', 'Y':
		return nil
	default:
		return errAborted
	}
}

// confirm prompts the user to confirm the question unless the --yes flag is
// set in cmd.
func confirm(cmd *cobra.Command, question string) error {
	yes, err := cmd.Flags().GetBool("yes")
	if yes || err != nil {
		return err
	}

	return ask(cmd.InOrStdin(), cmd.OutOrStdout(), question)
}
