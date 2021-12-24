package cmd

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cfs "github.com/charmbracelet/charm/fs"
	"github.com/rubiojr/tavern/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func init() {
	runtime.LockOSThread()
}

func TestPublish(t *testing.T) {
	buf := &testutil.Buffer{}
	log.SetOutput(buf)
	tdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := testutil.StartCharmServer(ctx, tdir)
	if err != nil {
		assert.FailNow(t, "error starting charm server")
	}

	cc, err := testutil.CharmClient()
	if err != nil {
		assert.FailNow(t, "error starting charm client")
	}

	cid, err := cc.ID()
	if err != nil {
		assert.FailNow(t, "error retrieving charm ID", err)
	}

	testutil.TavernServer(ctx, tdir)

	charmfs, err := cfs.NewFSWithClient(cc)
	if err != nil {
		assert.FailNow(t, "error creating charmfs client", err)
	}

	handle, err := os.Open("testdata/test.txt")
	assert.NoError(t, err)

	err = charmfs.WriteFile("testdata/test.txt", handle)
	assert.NoError(t, err)

	rootCmd.SetArgs([]string{
		"publish",
		"--charm-server-host", testutil.TestHost,
		"--server-url", testutil.ServerURL,
		"testdata/test.txt",
	})
	rootCmd.ExecuteC()
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(tdir, testutil.UploadsPath, cid, "testdata/test.txt"))
}
