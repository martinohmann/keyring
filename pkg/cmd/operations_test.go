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

	cmd := newRootCommand()
	cmd.SetOut(ioutil.Discard)
	cmd.SetArgs([]string{"get", testService, testUser})

	var buf bytes.Buffer

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

func TestSetCommand(t *testing.T) {
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
	require.Equal(secretSavedMsg+"\n", buf.String())

	secret, err := keyring.Get(testService, testUser)
	require.NoError(err)
	require.Equal(testSecret, secret)
}

func TestDeleteCommand(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	var buf bytes.Buffer

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser})
	cmd.SetOut(&buf)

	require.NoError(cmd.Execute())
	require.Equal(secretDeletedMsg+"\n", buf.String())

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
