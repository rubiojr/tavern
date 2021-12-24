package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/rubiojr/tavern/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestServe(t *testing.T) {
	buf := &testutil.Buffer{}
	log.SetOutput(buf)
	tdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd.SetArgs([]string{
		"serve",
		"--charm-server-url", "http://" + testutil.CharmServerHost + ":35354",
		"--path", tdir,
		"--address", testutil.TestServerURL,
	})
	go rootCmd.ExecuteContextC(ctx)
	testutil.WaitForServer("127.0.0.2:8000")

	assert.Regexp(t, regexp.MustCompile(`charm server: http://127.0.0.2:35354`), buf.String())
	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(`uploads directory: %s`, tdir)), buf.String())
	assert.Regexp(t, regexp.MustCompile(`serving on: 127.0.0.2:8000`), buf.String())
}
