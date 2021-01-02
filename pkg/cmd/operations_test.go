package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	keyring "github.com/zalando/go-keyring"
)

func TestGetCommand(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	cmd := newGetCommand()
	cmd.SetOut(ioutil.Discard)

	require.Error(cmd.Execute())

	cmd.SetArgs([]string{"myservice", "myuser"})

	require.Error(cmd.Execute())

	require.NoError(keyring.Set("myservice", "myuser", "supersecret"))

	var buf bytes.Buffer

	cmd.SetOut(&buf)
	require.NoError(cmd.Execute())
	require.Equal("supersecret\n", buf.String())
}

func TestSetCommand(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	r, w, err := os.Pipe()
	require.NoError(err)

	_, err = w.Write([]byte(`mypass`))
	require.NoError(err)
	require.NoError(w.Close())

	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	os.Stdin = r

	cmd := newSetCommand()
	cmd.SetOut(ioutil.Discard)

	require.Error(cmd.Execute())

	cmd.SetArgs([]string{"myservice", "myuser"})

	var buf bytes.Buffer

	cmd.SetOut(&buf)
	require.NoError(cmd.Execute())
	require.Equal(secretSavedMsg+"\n", buf.String())

	password, err := keyring.Get("myservice", "myuser")
	require.NoError(err)
	require.Equal("mypass", password)
}

func TestDeleteCommand(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	cmd := newDeleteCommand()
	cmd.SetOut(ioutil.Discard)

	require.Error(cmd.Execute())

	cmd.SetArgs([]string{"myservice", "myuser"})

	require.Error(cmd.Execute())

	require.NoError(keyring.Set("myservice", "myuser", "supersecret"))

	var buf bytes.Buffer

	cmd.SetOut(&buf)
	require.NoError(cmd.Execute())
	require.Equal(secretDeletedMsg+"\n", buf.String())
}
