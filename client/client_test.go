package client

import (
	"context"
	"os"
	"testing"

	cfs "github.com/charmbracelet/charm/fs"
	"github.com/rubiojr/tavern/internal/testutil"
	ts "github.com/rubiojr/tavern/server"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	tdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := testutil.StartCharmServer(ctx, tdir)

	tav := ts.NewServerWithConfig(&ts.Config{
		Addr:           "127.0.0.2:8000",
		UploadsPath:    tdir + "/uploads",
		CharmServerURL: "http://127.0.0.2:35354",
	})
	go tav.Serve(ctx)

	// Create a new client.
	tcc := DefaultConfig()
	tcc.CharmServerURL = testutil.CharmServerURL
	tcc.ServerURL = testutil.ServerURL
	c, err := NewClientWithConfig(tcc)
	if err != nil {
		assert.FailNow(t, "error creating tavern client")
	}

	cc, err := testutil.CharmClient()
	if err != nil {
		assert.FailNow(t, "error starting charm client")
	}

	charmfs, err := cfs.NewFSWithClient(cc)
	if err != nil {
		assert.FailNow(t, "error creating charmfs client", err)
	}

	handle, err := os.Open("testdata/test.txt")
	assert.NoError(t, err)

	err = charmfs.WriteFile("testdata/test.txt", handle)
	assert.NoError(t, err)

	err = c.Publish("testdata/test.txt")
	assert.NoError(t, err)
}
