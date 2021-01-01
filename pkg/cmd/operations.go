package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/term"
)

func newSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [service] [user]",
		Short: "Set password in keyring",
		Long: longDesc(`
			Sets a password for a service/user combination in the keyring.

			If stdin is a pipe, the password is read from there. Otherwise it will prompt for the password interactively.`),
		Example: example(`
		    # Password via stdin 
		    $ echo -n "supersecret" | keyring set myservice myuser

		    # Password via interactive prompt 
		    $ keyring set myservice myuser
		    Enter Password:`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			fi, err := os.Stdin.Stat()
			if err != nil {
				return err
			}

			var password []byte

			if (fi.Mode() & os.ModeCharDevice) == 0 {
				password, err = ioutil.ReadAll(os.Stdin)
			} else {
				fmt.Fprint(cmd.OutOrStdout(), "Enter Password: ")

				password, err = term.ReadPassword(int(os.Stdin.Fd()))
			}

			if err != nil {
				return err
			}

			err = keyring.Set(service, user, string(password))
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "password for service=%q,user=%q set\n", service, user)

			return nil
		},
	}

	return cmd
}

func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [service] [user]",
		Short: "Read password from keyring",
		Long: longDesc(`
			Reads a password for a service/user combination from the keyring and writes it to stdout.

			The returned password is terminated by a newline character.`),
		Example: example(`
		    $ keyring get myservice myuser`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			password, err := keyring.Get(service, user)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), password)

			return nil
		},
	}

	return cmd
}

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [service] [user]",
		Short: "Delete password from keyring",
		Long: longDesc(`
			Deletes the password for a service/user combination from the keyring.`),
		Example: example(`
		    $ keyring delete myservice myuser`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, user := args[0], args[1]

			err := keyring.Delete(service, user)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "password for service=%q,user=%q deleted\n", service, user)

			return nil
		},
	}

	return cmd
}
