package cmd

import (
	"bytes"
	"io/ioutil"
	"os"
	"runtime"
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

	cmd := newRootCommand()
	cmd.SetArgs([]string{"set", testService, testUser})
	cmd.SetIn(bytes.NewBuffer([]byte(testSecret)))
	cmd.SetOut(ioutil.Discard)

	require.NoError(cmd.Execute())

	secret, err := keyring.Get(testService, testUser)
	require.NoError(err)
	require.Equal(testSecret, secret)
}

func TestSetCommand_UpdateAutoConfirm(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	cmd := newRootCommand()
	cmd.SetArgs([]string{"set", testService, testUser, "--yes"})
	cmd.SetIn(bytes.NewBuffer([]byte(testSecret + "new")))
	cmd.SetOut(ioutil.Discard)

	require.NoError(cmd.Execute())

	secret, err := keyring.Get(testService, testUser)
	require.NoError(err)
	require.Equal(testSecret+"new", secret)
}

func TestSetCommand_UpdateAbort(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	cmd := newRootCommand()
	cmd.SetArgs([]string{"set", testService, testUser})
	cmd.SetIn(bytes.NewBuffer([]byte(testSecret + "new")))
	cmd.SetOut(ioutil.Discard)

	require.Error(cmd.Execute())

	secret, err := keyring.Get(testService, testUser)
	require.NoError(err)
	require.Equal(testSecret, secret)
}

func TestDeleteCommand_Abort(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser})
	cmd.SetIn(bytes.NewBuffer(nil))
	cmd.SetOut(ioutil.Discard)

	require.Equal(errConfirmationRequired, cmd.Execute())

	_, err := keyring.Get(testService, testUser)
	require.NoError(err)
}

func TestDeleteCommand_AutoConfirm(t *testing.T) {
	require := require.New(t)

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser, "--yes"})
	cmd.SetOut(ioutil.Discard)

	require.NoError(cmd.Execute())

	_, err := keyring.Get(testService, testUser)
	require.Error(err)
}

func TestDeleteCommand_UserAbort(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
	}

	require := require.New(t)

	ptmx, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	require.NoError(err)
	defer ptmx.Close()

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	_, err = ptmx.Write([]byte(`n`))
	require.NoError(err)

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser})
	cmd.SetIn(ptmx)
	cmd.SetOut(ioutil.Discard)

	require.Equal(errAborted, cmd.Execute())

	_, err = keyring.Get(testService, testUser)
	require.NoError(err)
}

func TestDeleteCommand_UserConfirm(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skipf("unknown terminal path for GOOS %v", runtime.GOOS)
	}

	require := require.New(t)

	ptmx, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	require.NoError(err)
	defer ptmx.Close()

	keyring.MockInit()

	require.NoError(keyring.Set(testService, testUser, testSecret))

	_, err = ptmx.Write([]byte(`y`))
	require.NoError(err)

	cmd := newRootCommand()
	cmd.SetArgs([]string{"delete", testService, testUser})
	cmd.SetIn(ptmx)
	cmd.SetOut(ioutil.Discard)

	require.NoError(cmd.Execute())

	_, err = keyring.Get(testService, testUser)
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
