package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
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
