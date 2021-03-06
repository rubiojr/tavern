package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfs "github.com/charmbracelet/charm/fs"
	"github.com/rubiojr/tavern/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	cc, err := testutil.CharmClient()
	if err != nil {
		assert.FailNow(t, "error starting charm client", err)
	}

	cid, err := cc.ID()
	if err != nil {
		assert.FailNow(t, "error retrieving charm ID", err)
	}

	charmfs, err := cfs.NewFSWithClient(cc)
	if err != nil {
		assert.FailNow(t, "error creating charmfs client", err)
	}

	handle, err := os.Open("testdata/test.txt")
	assert.NoError(t, err)

	err = charmfs.WriteFile("testdata/test.txt", handle)
	assert.NoError(t, err)

	t.Run("publish testdata/test.txt", func(t *testing.T) {
		tdir := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = testutil.TavernServer(ctx, tdir)
		assert.NoError(t, err)

		rootCmd.SetArgs([]string{
			"publish",
			"--charm-server-host", testutil.CharmServerHost,
			"--server-url", testutil.TestServerURL,
			"testdata/test.txt",
		})
		_, err = rootCmd.ExecuteC()
		assert.NoError(t, err)
		dfile := filepath.Join(tdir, testutil.UploadsPath, cid, "testdata/test.txt")
		assert.FileExists(t, dfile)

		out, err := ioutil.ReadFile(dfile)
		assert.NoError(t, err)
		assert.Equal(t, "foo", strings.TrimRight(string(out), "\r\n"))
	})

	t.Run("publishing not allowed", func(t *testing.T) {
		tdir := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = testutil.TavernServerA(ctx, tdir, "foo.bar")
		assert.NoError(t, err)

		rootCmd.SetArgs([]string{
			"publish",
			"--charm-server-host", testutil.CharmServerHost,
			"--server-url", testutil.TestServerURL,
			"testdata/test.txt",
		})
		_, err = rootCmd.ExecuteC()
		assert.EqualError(t, err, "publishing failed: {\"error\":\"charm server localhost cannot publish\"}")
	})
}
