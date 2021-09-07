package core

import (
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/rpc"
)

// Config combines all configuration fields for Core subsystem.
type Config struct {
	EnableRemote bool
	Remote       struct {
		Protocol   string
		RemoteAddr string
	}
	EmbeddedConfig *core.Config
}

// DefaultConfig returns default configuration for Core subsystem.
func DefaultConfig() *Config {
	return &Config{
		EmbeddedConfig: core.DefaultConfig(),
		EnableRemote:   false,
	}
}

// Components collects all the components and services related to the Core node.
func Components(cfg *Config) fx.Option {
	return fx.Options(
		fx.Provide(rpc.NewClient),
		fx.Provide(func() (core.Client, error) {
			if cfg.EnableRemote {
				return core.NewRemote(cfg.Remote.Protocol, cfg.Remote.RemoteAddr)
			}

			return core.NewEmbedded(cfg.EmbeddedConfig)
		}),
	)
}
