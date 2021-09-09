package core

import (
	"fmt"

	"github.com/celestiaorg/celestia-core/config"
	corenode "github.com/celestiaorg/celestia-core/node"
	"github.com/celestiaorg/celestia-core/rpc/client"
	"github.com/celestiaorg/celestia-core/rpc/client/http"
	"github.com/celestiaorg/celestia-core/rpc/client/local"
)

// Client is an alias to Core Client.
type Client = client.Client

// NewRemote creates a new Client that communicates with a remote Core endpoint over HTTP.
func NewRemote(protocol, remoteAddr string) (Client, error) {
	return http.New(
		fmt.Sprintf("%s://%s", protocol, remoteAddr),
		"/websocket",
	)
}

// NewEmbedded creates a new Client which works over embedded Core in the same proc.
func NewEmbedded(cfg *config.Config) (Client, error) {
	node, err := corenode.DefaultNewNode(cfg, adaptedLogger())
	if err != nil {
		return nil, err
	}

	return &embeddedWrapper{local.New(node), node}, nil
}

// NewEmbeddedFromNode wraps embeds a given Core Node to control its lifecycle.
func NewEmbeddedFromNode(node *corenode.Node) Client {
	return &embeddedWrapper{local.New(node), node}
}

// embeddedWrapper is a small wrapper around local Client which ensures the embedded Core node is started/stopped.
type embeddedWrapper struct {
	*local.Local
	node *corenode.Node
}

func (e *embeddedWrapper) Start() error {
	return e.node.Start()
}

func (e *embeddedWrapper) Stop() error {
	return e.node.Stop()
}
