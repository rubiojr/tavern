package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/charm/client"
	cfs "github.com/charmbracelet/charm/fs"
	"github.com/charmbracelet/charm/server"
	"github.com/charmbracelet/keygen"
	ts "github.com/rubiojr/tavern/server"
	"github.com/stretchr/testify/assert"
)

func TestPublish(t *testing.T) {
	tdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := server.DefaultConfig()
	cfg.DataDir = tdir

	sp := fmt.Sprintf("%s/.ssh", cfg.DataDir)
	kp, err := keygen.NewWithWrite(sp, "charm_server", []byte(""), keygen.RSA)
	if err != nil {
		assert.FailNow(t, "error generating server keys")
	}
	cfg.WithKeys(kp.PublicKey, kp.PrivateKeyPEM)

	charm, err := server.NewServer(cfg)
	if err != nil {
		panic(err)
	}
	go charm.Start(ctx)

	if !waitForServer(":35354") {
		t.Fatal("charm server did not start")
	}

	tav := ts.NewServerWithConfig(&ts.Config{
		Addr:           "127.0.0.2:8000",
		UploadsPath:    tdir + "/uploads",
		CharmServerURL: "http://127.0.0.2:35354",
	})
	go tav.Serve(ctx)

	// Create a new client.
	tcc := DefaultConfig()
	tcc.CharmServerURL = "http://127.0.0.2"
	tcc.ServerURL = "http://127.0.0.2:8000"
	c, err := NewClientWithConfig(tcc)
	if err != nil {
		t.Fatal(err)
	}

	// Charm inits
	err = genClientKeys("127.0.0.2")
	assert.NoError(t, err)

	cconfig, err := client.ConfigFromEnv()
	if err != nil {
		assert.FailNow(t, "error loading charm client config")
	}
	cconfig.Host = "127.0.0.2"

	cc, err := client.NewClient(cconfig)
	assert.NoError(t, err)
	cc.Config.SSHPort = 35353
	cc.Config.HTTPPort = 35354

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

func waitForServer(addr string) bool {
	for i := 0; i < 40; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}

	return false
}

func genClientKeys(dir string) error {
	// Generate keys
	dp, err := client.DataPath(dir)
	if err != nil {
		return err
	}
	_, err = keygen.NewWithWrite(dp, "charm", []byte(""), keygen.RSA)
	if err != nil {
		return err
	}

	return nil
}
