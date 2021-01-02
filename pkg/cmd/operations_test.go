package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	keyring "github.com/zalando/go-keyring"
)

const (
	testService = "the-service"
	testUser    = "the-user"
	testSecret  = "the-secret"
)

func TestGetCommand(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"get", testService, testUser})
	cmd.SetOut(&buf)

	require.NoError(cmd.Execute())
	require.Equal(testSecret, buf.String())
}

func TestGetCommand_NotFound(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	cmd := newRootCommand()
	cmd.SetOut(ioutil.Discard)
	cmd.SetArgs([]string{"get", testService, testUser})

	require.Error(cmd.Execute())
}

func TestSetCommand_Create(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	_, err := keyring.Get(testService, testUser)
	require.Error(err)

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"set", testService, testUser})
	cmd.SetIn(bytes.NewBuffer([]byte(testSecret)))
	cmd.SetOut(&buf)

	require.NoError(cmd.Execute())
	require.Contains(buf.String(), secretCreatedMsg)

	secret, err := keyring.Get(testService, testUser)
	require.NoError(err)
	require.Equal(testSecret, secret)
}

func TestSetCommand_UpdateConfirm(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"set", testService, testUser, "--yes"})
	cmd.SetIn(bytes.NewBuffer([]byte(testSecret + "new")))
	cmd.SetOut(&buf)

	require.NoError(cmd.Execute())
	require.Contains(buf.String(), secretUpdatedMsg)

	secret, err := keyring.Get(testService, testUser)
	require.NoError(err)
	require.Equal(testSecret+"new", secret)
}

func TestSetCommand_UpdateAbort(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"set", testService, testUser})
	cmd.SetIn(bytes.NewBuffer([]byte(testSecret + "new")))
	cmd.SetOut(&buf)

	require.Error(cmd.Execute())

	secret, err := keyring.Get(testService, testUser)
	require.NoError(err)
	require.Equal(testSecret, secret)
}

func TestDeleteCommand_Confirm(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser})
	cmd.SetIn(bytes.NewBuffer([]byte(`y`)))
	cmd.SetOut(&buf)

	require.NoError(cmd.Execute())
	require.Contains(buf.String(), secretDeletedMsg)

	_, err := keyring.Get(testService, testUser)
	require.Error(err)
}

func TestDeleteCommand_Abort(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser})
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(&buf)

	require.Error(cmd.Execute())

	_, err := keyring.Get(testService, testUser)
	require.NoError(err)
}

func TestDeleteCommand_AutoConfirm(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser, "--yes"})
	cmd.SetOut(&buf)

	require.NoError(cmd.Execute())
	require.Contains(buf.String(), secretDeletedMsg)

	_, err := keyring.Get(testService, testUser)
	require.Error(err)
}

func TestDeleteCommand_NotFound(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	cmd := newRootCommand()
	cmd.SetOut(ioutil.Discard)
	cmd.SetArgs([]string{"delete", testService, testUser})

	require.Error(cmd.Execute())
}
