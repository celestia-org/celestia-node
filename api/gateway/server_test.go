package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	address, port := "localhost", "0"
	server := NewServer(address, port)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	err := server.Start(ctx)
	require.NoError(t, err)

	// register ping handler
	ping := new(ping)
	server.RegisterHandlerFunc("/ping", ping.ServeHTTP, http.MethodGet)

	url := fmt.Sprintf("http://%s/ping", server.ListenAddr())

	resp, err := http.Get(url)
	require.NoError(t, err)

	buf, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	t.Cleanup(func() {
		resp.Body.Close()
	})
	assert.Equal(t, "pong", string(buf))

	err = server.Stop(ctx)
	require.NoError(t, err)
}

type ping struct{}

func (p ping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//nolint:errcheck
	w.Write([]byte("pong"))
}
