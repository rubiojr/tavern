package testutil

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/charm/client"
	"github.com/charmbracelet/charm/server"
	"github.com/charmbracelet/keygen"
	ts "github.com/rubiojr/tavern/server"
)

const TestHost = "127.0.0.2"
const TestServerURL = TestHost + ":8000"
const CharmServerHost = TestHost
const ServerURL = "http://" + TestHost + ":8000"
const UploadsPath = "/uploads"

// Thread safe buffer to avoid data races when setting a custom writer
// for the log
type Buffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}

func (b *Buffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func StartCharmServer(ctx context.Context, dataDir string) error {
	cfg := server.DefaultConfig()
	cfg.DataDir = dataDir

	sp := fmt.Sprintf("%s/.ssh", cfg.DataDir)
	kp, err := keygen.NewWithWrite(sp, "charm_server", []byte(""), keygen.RSA)
	if err != nil {
		return err
	}
	cfg.WithKeys(kp.PublicKey, kp.PrivateKeyPEM)

	charm, err := server.NewServer(cfg)
	if err != nil {
		return err
	}
	go charm.Start(ctx)

	if !WaitForServer(":35354") {
		return fmt.Errorf("charm server did not start")
	}

	return nil
}

func CharmClient() (*client.Client, error) {
	err := genClientKeys("127.0.0.2")
	if err != nil {
		return nil, err
	}

	cconfig, err := client.ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	cconfig.Host = "127.0.0.2"

	cc, err := client.NewClient(cconfig)
	if err != nil {
		return nil, err
	}

	cc.Config.SSHPort = 35353
	cc.Config.HTTPPort = 35354

	return cc, nil
}

func TavernServer(ctx context.Context, dataDir string) (*ts.Server, error) {
	tav := ts.NewServerWithConfig(&ts.Config{
		Addr:           "127.0.0.2:8000",
		UploadsPath:    filepath.Join(dataDir, UploadsPath),
		CharmServerURL: "http://127.0.0.2:35354",
	})
	go tav.Serve(ctx)

	if !WaitForServer("127.0.0.2:8000") {
		return nil, fmt.Errorf("tavern server did not start")
	}

	return tav, nil
}

func WaitForServer(addr string) bool {
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
