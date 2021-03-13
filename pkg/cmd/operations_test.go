package cmd

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/martinohmann/exit"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	keyring "github.com/zalando/go-keyring"
)

const (
	testService = "the-service"
	testUser    = "the-user"
	testSecret  = "the-secret"
)

var testCases = []struct {
	name        string
	args        []string
	stdin       io.Reader
	stdout      io.Writer
	expectedErr error
	setup       func(t *testing.T, cmd *cobra.Command) (cleanup func())
	assert      func(t *testing.T, cmd *cobra.Command)
}{
	{
		name:   "read existing secret",
		args:   []string{"get", testService, testUser},
		stdout: bytes.NewBuffer(nil),
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			createSecret(t, testService, testUser, testSecret)
			return nil
		},
		assert: func(t *testing.T, cmd *cobra.Command) {
			buf := cmd.OutOrStdout().(*bytes.Buffer)
			require.Equal(t, testSecret, buf.String())
		},
	},
	{
		name:        "read existing secret, bad out -> error",
		args:        []string{"get", testService, testUser},
		stdout:      badWriter{},
		expectedErr: exit.Error(exit.CodeIOErr, io.ErrClosedPipe),
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			createSecret(t, testService, testUser, testSecret)
			return nil
		},
	},
	{
		name:        "read nonexistent secret -> error",
		args:        []string{"get", testService, testUser},
		expectedErr: keyring.ErrNotFound,
	},
	{
		name:  "create secret",
		args:  []string{"set", testService, testUser},
		stdin: bytes.NewBuffer([]byte(testSecret)),
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretEquals(t, testService, testUser, testSecret)
		},
	},
	{
		name: "create secret interactively",
		args: []string{"set", testService, testUser},
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			if runtime.GOOS != "linux" {
				t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
			}

			ptmx, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
			require.NoError(t, err)

			cmd.SetIn(ptmx)

			// write secret to stdin
			_, err = ptmx.Write([]byte(testSecret + "\r\n"))
			require.NoError(t, err)

			return func() { ptmx.Close() }
		},
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretEquals(t, testService, testUser, testSecret)
		},
	},
	{
		name:        "stdin broken pipe while creating secret -> error",
		args:        []string{"set", testService, testUser},
		stdin:       badReader{},
		expectedErr: exit.Error(exit.CodeIOErr, io.ErrClosedPipe),
	},
	{
		name:  "update existing secret from stdin with --yes flag",
		args:  []string{"set", testService, testUser, "--yes"},
		stdin: bytes.NewBuffer([]byte(testSecret + "new")),
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			createSecret(t, testService, testUser, testSecret)
			return nil
		},
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretEquals(t, testService, testUser, testSecret+"new")
		},
	},
	{
		name:  "update existing secret from stdin without --yes flag -> error",
		args:  []string{"set", testService, testUser},
		stdin: bytes.NewBuffer([]byte(testSecret + "new")),
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			createSecret(t, testService, testUser, testSecret)
			return nil
		},
		expectedErr: exit.Error(exit.CodeUsage, errConfirmationRequired),
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretEquals(t, testService, testUser, testSecret)
		},
	},
	{
		name: "delete existing secret with --yes flag",
		args: []string{"delete", testService, testUser, "--yes"},
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			createSecret(t, testService, testUser, testSecret)
			return nil
		},
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretNotExists(t, testService, testUser)
		},
	},
	{
		name: "delete existing secret without --yes flag -> error",
		args: []string{"delete", testService, testUser},
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			createSecret(t, testService, testUser, testSecret)
			return nil
		},
		expectedErr: exit.Error(exit.CodeUsage, errConfirmationRequired),
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretEquals(t, testService, testUser, testSecret)
		},
	},
	{
		name: "delete existing secret interactively, user aborts",
		args: []string{"delete", testService, testUser},
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			if runtime.GOOS != "linux" {
				t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
			}

			ptmx, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
			require.NoError(t, err)

			createSecret(t, testService, testUser, testSecret)

			cmd.SetIn(ptmx)

			// reject deletion prompt
			_, err = ptmx.Write([]byte(`n`))
			require.NoError(t, err)

			return func() { ptmx.Close() }
		},
		expectedErr: errAborted,
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretEquals(t, testService, testUser, testSecret)
		},
	},
	{
		name: "delete existing secret interactively, user confirms",
		args: []string{"delete", testService, testUser},
		setup: func(t *testing.T, cmd *cobra.Command) func() {
			if runtime.GOOS != "linux" {
				t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
			}

			ptmx, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
			require.NoError(t, err)

			createSecret(t, testService, testUser, testSecret)

			cmd.SetIn(ptmx)

			// confirm deletion prompt
			_, err = ptmx.Write([]byte(`y`))
			require.NoError(t, err)

			return func() { ptmx.Close() }
		},
		assert: func(t *testing.T, cmd *cobra.Command) {
			assertSecretNotExists(t, testService, testUser)
		},
	},
	{
		name:        "delete nonexistent secret -> error",
		args:        []string{"delete", testService, testUser},
		expectedErr: keyring.ErrNotFound,
	},
}

func TestOperations(t *testing.T) {
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)

			keyring.MockInit()

			cmd := newRootCommand()
			cmd.SetArgs(test.args)
			cmd.SetOut(ioutil.Discard)

			if test.stdin != nil {
				cmd.SetIn(test.stdin)
			}

			if test.stdout != nil {
				cmd.SetOut(test.stdout)
			}

			if test.setup != nil {
				if cleanup := test.setup(t, cmd); cleanup != nil {
					defer cleanup()
				}
			}

			err := cmd.Execute()
			if test.expectedErr == nil {
				require.NoError(err)
			} else {
				require.EqualError(err, test.expectedErr.Error())
			}

			if test.assert != nil {
				test.assert(t, cmd)
			}
		})
	}
}

func createSecret(t *testing.T, svc, user, pass string) {
	require.NoError(t, keyring.Set(svc, user, pass))
}

func assertSecretNotExists(t *testing.T, svc, user string) {
	_, err := keyring.Get(svc, user)
	require.Equal(t, keyring.ErrNotFound, err)
}

func assertSecretEquals(t *testing.T, svc, user, expected string) {
	actual, err := keyring.Get(svc, user)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

type badReader struct{}

func (badReader) Read(_ []byte) (int, error) {
	return 0, io.ErrClosedPipe
}

type badWriter struct{}

func (badWriter) Write(_ []byte) (int, error) {
	return 0, io.ErrClosedPipe
}
