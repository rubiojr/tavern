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
	serverAddr := "127.0.0.1:8001"

	t.Run("test serve", func(t *testing.T) {
		buf := &testutil.Buffer{}
		log.SetOutput(buf)
		tdir := t.TempDir()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		rootCmd.SetArgs([]string{
			"serve",
			"--path", tdir,
			"--address", serverAddr,
		})
		go rootCmd.ExecuteContextC(ctx)

		if !testutil.WaitForServer(serverAddr) {
			assert.FailNow(t, "error starting tavern server")
		}

		assert.True(t, strings.Contains(buf.String(), fmt.Sprintf(`uploads directory: %s`, tdir)))
		assert.Regexp(t, regexp.MustCompile(`serving on: `+serverAddr), buf.String())
	})

	t.Run("invalid JWT", func(t *testing.T) {
		buf := &testutil.Buffer{}
		log.SetOutput(buf)
		tdir := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rootCmd.SetArgs([]string{
			"serve",
			"--path", tdir,
			"--address", serverAddr,
		})
		go rootCmd.ExecuteContextC(ctx)

		if !testutil.WaitForServer(serverAddr) {
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
	})
	t.Run("valid JWT", func(t *testing.T) {
		buf := &testutil.Buffer{}
		log.SetOutput(buf)
		tdir := t.TempDir()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		rootCmd.SetArgs([]string{
			"serve",
			"--path", tdir,
			"--address", serverAddr,
		})
		go rootCmd.ExecuteContextC(ctx)

		if !testutil.WaitForServer(serverAddr) {
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
	})
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
	cfg.ServerURL = "http://127.0.0.1:8001"
	cfg.CharmServerHost = testutil.CharmServerHost
	return client.NewClientWithConfig(cfg)
}
