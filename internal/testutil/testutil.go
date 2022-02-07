package testutil

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/charm/client"
	"github.com/charmbracelet/keygen"
	ts "github.com/rubiojr/tavern/server"
)

const TestHost = "127.0.0.2"
const TestServerAddr = TestHost + ":8000"
const CharmServerHost = "127.0.0.1"
const TestServerURL = "http://" + TestHost + ":8000"
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

func CharmClient() (*client.Client, error) {
	err := genClientKeys(TestHost)
	if err != nil {
		return nil, err
	}

	cconfig, err := client.ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	cconfig.Host = TestHost

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
		Addr:        TestServerAddr,
		UploadsPath: filepath.Join(dataDir, UploadsPath),
	})
	go tav.Serve(ctx)

	if !WaitForServer(TestServerAddr) {
		return nil, fmt.Errorf("tavern server did not start")
	}

	return tav, nil
}

// Start a Tavern server with an allowed list of Charm servers
func TavernServerA(ctx context.Context, dataDir string, allowList ...string) (*ts.Server, error) {
	tav := ts.NewServerWithConfig(&ts.Config{
		Addr:                TestServerAddr,
		UploadsPath:         filepath.Join(dataDir, UploadsPath),
		AllowedCharmServers: allowList,
	})
	go tav.Serve(ctx)

	if !WaitForServer(TestServerAddr) {
		return nil, fmt.Errorf("tavern server did not start")
	}

	return tav, nil
}

func WaitForServerShutdown(addr string) bool {
	for i := 0; i < 40; i++ {
		_, err := net.Dial("tcp", addr)
		if err != nil {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}

	return false
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

func init() {
	// Ugly hack until we can set path for keys via Env in Charm
	// https://github.com/charmbracelet/charm/issues/50
	os.Setenv("HOME", filepath.Join("../_fixtures/home"))
	// for Windows tests
	os.Setenv("LOCALAPPDATA", filepath.Join("../_fixtures/home/.local/share"))
	os.Setenv("CHARM_HTTP_SCHEME", "http")
	fmt.Println("foooo")
}
