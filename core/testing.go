package core

import (
	"github.com/tendermint/tendermint/abci/example/kvstore"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/node"
	rpctest "github.com/tendermint/tendermint/rpc/test"
)

const defaultRetainBlocks int64 = 10

// StartMockNode starts a mock Core node background process and returns it.
func StartMockNode(app types.Application) *node.Node {
	return rpctest.StartTendermint(app, rpctest.SuppressStdout, rpctest.RecreateConfig)
}

// CreateKvStore creates a simple kv store app and gives the user
// ability to set desired amount of blocks to be retained.
func CreateKvStore(retainBlocks int64) *kvstore.Application {
	app := kvstore.NewApplication()
	app.RetainBlocks = retainBlocks
	return app
}

// StartRemoteClient returns a started remote Core node process, as well its
// mock Core Client.
func StartRemoteClient() (*node.Node, Client, error) {
	remote := StartMockNode(CreateKvStore(defaultRetainBlocks))
	protocol, ip := GetRemoteEndpoint(remote)
	client, err := NewRemote(protocol, ip)
	return remote, client, err
}

// GetRemoteEndpoint returns the protocol and ip of the remote node.
func GetRemoteEndpoint(remote *node.Node) (string, string) {
	endpoint := remote.Config().RPC.ListenAddress
	// protocol = "tcp"
	protocol, ip := endpoint[:3], endpoint[6:]
	return protocol, ip
}
