package node

import (
	"context"

	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/node/config"
	"github.com/celestiaorg/celestia-node/node/p2p"
)

// NewFull creates and runs a new ready-to-go Full Node.
// To gracefully stop it the Stop method must be used.
func NewFull(ctx context.Context, cfg *config.Config) (*Node, error) {
	node, err := newNode(full(cfg))
	if err != nil {
		return nil, err
	}

	err = node.Start(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: Add bootstrapping
	return node, nil
}

// full keeps all the DI options required to built a Full Node.
func full(cfg *config.Config) fx.Option {
	return fx.Options(
		fx.Provide(func() Type {
			return Full
		}),
		fx.Provide(func() *config.Config {
			return cfg
		}),
		fx.Provide(p2p.Host),
	)
}
