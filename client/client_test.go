package client

import (
	"context"
	"os"
	"testing"

	cfs "github.com/charmbracelet/charm/fs"
	"github.com/rubiojr/tavern/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
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

	testutil.TavernServer(ctx, tdir)

	// Create a new client.
	tcc := DefaultConfig()
	tcc.CharmServerURL = testutil.CharmServerURL
	tcc.ServerURL = testutil.ServerURL
	c, err := NewClientWithConfig(tcc)
	if err != nil {
		assert.FailNow(t, "error creating tavern client", err)
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
