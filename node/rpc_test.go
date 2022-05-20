package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/header/local"
	"github.com/celestiaorg/celestia-node/header/store"
	"github.com/celestiaorg/celestia-node/header/sync"
	service "github.com/celestiaorg/celestia-node/service/header"
	"github.com/celestiaorg/celestia-node/service/rpc"
)

// NOTE: The following tests are against common RPC endpoints provided by
// celestia-node. They will be removed upon refactoring of the RPC
// architecture and Public API. @renaynay @Wondertan.

// TestNamespacedSharesRequest tests the `/namespaced_shares` endpoint.
func TestNamespacedSharesRequest(t *testing.T) {
	nd := setupNodeWithModifiedRPC(t)
	// create several requests for header at height 2
	height := uint64(2)
	var tests = []struct {
		nID         string
		expectedErr bool
		errMsg      string
	}{
		{
			nID:         "0000000000000001",
			expectedErr: false,
		},
		{
			nID:         "00000000000001",
			expectedErr: true,
			errMsg:      "expected namespace ID of size 8, got 7",
		},
		{
			nID:         "000000000000000001",
			expectedErr: true,
			errMsg:      "expected namespace ID of size 8, got 9",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			endpoint := fmt.Sprintf("http://127.0.0.1:%s/namespaced_shares/%s/height/%d",
				nd.RPCServer.ListenAddr()[5:], tt.nID, height)
			resp, err := http.Get(endpoint)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()
			// check resp
			if tt.expectedErr {
				require.False(t, resp.StatusCode == http.StatusOK)

				buf, err := ioutil.ReadAll(resp.Body)
				require.NoError(t, err)
				expectedErr := strings.Trim(string(buf), "\"\"") //nolint:staticcheck
				require.Equal(t, tt.errMsg, expectedErr)

				return
			}
			require.NoError(t, err)
			require.True(t, resp.StatusCode == http.StatusOK)
			// decode resp
			namespacedShares := new(rpc.NamespacedSharesResponse)
			err = json.NewDecoder(resp.Body).Decode(namespacedShares)
			require.Equal(t, height, namespacedShares.Height)
		})
	}
}

// TestHeadRequest rests the `/head` endpoint.
func TestHeadRequest(t *testing.T) {
	nd := setupNodeWithModifiedRPC(t)
	endpoint := fmt.Sprintf("http://127.0.0.1:%s/head", nd.RPCServer.ListenAddr()[5:])
	resp, err := http.Get(endpoint)
	require.NoError(t, err)
	defer func() {
		err = resp.Body.Close()
		require.NoError(t, err)
	}()
	require.True(t, resp.StatusCode == http.StatusOK)
}

// TestHeaderRequest tests the `/header` endpoint.
func TestHeaderRequest(t *testing.T) {
	nd := setupNodeWithModifiedRPC(t)
	// create several requests for headers
	var tests = []struct {
		height      uint64
		expectedErr bool
	}{
		{
			height:      uint64(2),
			expectedErr: false,
		},
		{
			height:      uint64(0),
			expectedErr: true,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			endpoint := fmt.Sprintf("http://127.0.0.1:%s/header/%d", nd.RPCServer.ListenAddr()[5:], tt.height)
			resp, err := http.Get(endpoint)
			require.NoError(t, err)
			defer func() {
				err = resp.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tt.expectedErr, resp.StatusCode != http.StatusOK)
		})
	}
}

func setupNodeWithModifiedRPC(t *testing.T) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	// create test node with a dummy header service, manually add a dummy header
	// service and register it with rpc handler/server
	hServ := setupHeaderService(ctx, t)
	// create overrides
	overrideHeaderServ := func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(hServ))
	}
	overrideRPCHandler := func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(func(srv *rpc.Server) *rpc.Handler {
			handler := rpc.NewHandler(nil, nil, hServ)
			handler.RegisterEndpoints(srv)
			return handler
		}))
	}
	nd := TestNode(t, Full, overrideHeaderServ, overrideRPCHandler)
	// start node
	err := nd.Start(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = nd.Stop(ctx)
		require.NoError(t, err)
	})
	return nd
}

func setupHeaderService(ctx context.Context, t *testing.T) *service.Service {
	suite := header.NewTestSuite(t, 1)
	head := suite.Head()
	// create header stores
	remoteStore := store.NewTestStore(ctx, t, head)
	localStore := store.NewTestStore(ctx, t, head)
	_, err := localStore.Append(ctx, suite.GenExtendedHeaders(5)...)
	require.NoError(t, err)
	// create syncer
	syncer := sync.NewSyncer(local.NewExchange(remoteStore), localStore, &header.DummySubscriber{})

	return service.NewHeaderService(syncer, nil, nil, nil, localStore)
}
