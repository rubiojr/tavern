package cmd

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/charm/server"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rubiojr/tavern/client"
	"github.com/rubiojr/tavern/internal/testutil"
	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

const charmID = "b4ede63d-c736-4561-80e9-0f912337b251"

func TestServe(t *testing.T) {
	buf := &testutil.Buffer{}
	log.SetOutput(buf)
	tdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd.SetArgs([]string{
		"serve",
		"--path", tdir,
		"--address", testutil.TestServerURL,
	})
	go rootCmd.ExecuteContextC(ctx)

	if !testutil.WaitForServer("127.0.0.2:8000") {
		assert.FailNow(t, "error starting tavern server")
	}

	assert.True(t, strings.Contains(buf.String(), fmt.Sprintf(`uploads directory: %s`, tdir)))
	assert.Regexp(t, regexp.MustCompile(`serving on: 127.0.0.2:8000`), buf.String())
}

func TestInvalidJWT(t *testing.T) {
	buf := &testutil.Buffer{}
	log.SetOutput(buf)
	tdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd.SetArgs([]string{
		"serve",
		"--path", tdir,
		"--address", testutil.TestServerURL,
	})
	go rootCmd.ExecuteContextC(ctx)

	if !testutil.WaitForServer("127.0.0.2:8000") {
		assert.FailNow(t, "error starting tavern server")
	}

	pk, err := generateEd25519Keys()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, &jwt.RegisteredClaims{
		Issuer:    "http://localhost:35354",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Audience:  []string{"tavern"},
		Subject:   charmID,
	})
	sum := sha256.Sum256(*pk)
	kid := fmt.Sprintf("%x", sum)
	token.Header["kid"] = kid
	// Ugly hack until we can set path for keys via Env in Charm
	// https://github.com/charmbracelet/charm/issues/50
	os.Setenv("HOME", filepath.Join("../_fixtures/home"))
	// for Windows tests
	os.Setenv("LOCALAPPDATA", filepath.Join("../_fixtures/home/.local/share"))
	tokenString, err := token.SignedString(pk)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	c, err := tClient()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	bbuf := bytes.NewBuffer([]byte{})
	req, err := c.UploadRequest(tokenString, bbuf)
	assert.NoError(t, err)
	httpc := &http.Client{}
	resp, err := httpc.Do(req)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, 401, resp.StatusCode)
	assert.Regexp(t, regexp.MustCompile(`JWT validation failed`), buf.String())
}

func TestValidJWT(t *testing.T) {
	buf := &testutil.Buffer{}
	log.SetOutput(buf)
	tdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd.SetArgs([]string{
		"serve",
		"--path", tdir,
		"--address", testutil.TestServerURL,
	})
	go rootCmd.ExecuteContextC(ctx)

	if !testutil.WaitForServer("127.0.0.2:8000") {
		assert.FailNow(t, "error starting tavern server")
	}

	rpk, err := ioutil.ReadFile("../_fixtures/server/.ssh/charm_server_ed25519")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	pk, err := gossh.ParseRawPrivateKey(rpk)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	jwkp := server.NewJSONWebKeyPair(pk.(*ed25519.PrivateKey))

	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, &jwt.RegisteredClaims{
		Issuer:    "http://localhost:35354",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Audience:  []string{"tavern"},
		Subject:   charmID,
	})
	sum := sha256.Sum256([]byte(*jwkp.PrivateKey))
	kid := fmt.Sprintf("%x", sum)
	token.Header["kid"] = kid
	// Ugly hack until we can set path for keys via Env in Charm
	// https://github.com/charmbracelet/charm/issues/50
	os.Setenv("HOME", filepath.Join("../_fixtures/home"))
	// for Windows tests
	os.Setenv("LOCALAPPDATA", filepath.Join("../_fixtures/home/.local/share"))
	tokenString, err := token.SignedString(pk)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	c, err := tClient()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	req, err := c.UploadRequest(tokenString, bytes.NewBuffer([]byte{}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	httpc := &http.Client{}
	resp, err := httpc.Do(req)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	rout, err := io.ReadAll(resp.Body)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "no files found in request\n", string(rout))
}

// generateEd25519Keys creates a pair of EdD25519 keys for SSH auth.
func generateEd25519Keys() (*ed25519.PrivateKey, error) {
	// Generate keys
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &privateKey, nil
}

func tClient() (*client.Client, error) {
	cfg := client.DefaultConfig()
	cfg.ServerURL = testutil.ServerURL
	cfg.CharmServerHost = testutil.CharmServerHost
	return client.NewClientWithConfig(cfg)
}
