package config

import (
	"io"

	"github.com/pelletier/go-toml/v2"

	"github.com/celestiaorg/celestia-node/node/core"
	"github.com/celestiaorg/celestia-node/node/p2p"
)

// Config is main configuration structure for a Node.
// It combines configuration units for all Node subsystems.
type Config struct {
	P2P  p2p.Config
	Core core.Config
}

// DefaultConfig provides a default Node Config.
func DefaultConfig() *Config {
	return &Config{
		P2P:  p2p.DefaultConfig(),
		Core: core.DefaultConfig(),
	}
}

// TODO(@Wondertan): We should have a description for each field written into w,
//  so users can instantly understand purpose of each field. Ideally, we should have a utility program to parse comments
//  from actual sources(*.go files) and generate docs from comments. Hint: use 'ast' package.
// WriteTo flushes a given Config into w.
func WriteTo(cfg *Config, w io.Writer) error {
	return toml.NewEncoder(w).Encode(cfg)
}

// ReadFrom pulls a Config from a given reader r.
func ReadFrom(r io.Reader) (*Config, error) {
	var cfg Config
	return &cfg, toml.NewDecoder(r).Decode(&cfg)
}
