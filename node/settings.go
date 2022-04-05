package node

import (
	"encoding/hex"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"go.uber.org/fx"

	"github.com/celestiaorg/celestia-node/core"
	"github.com/celestiaorg/celestia-node/node/p2p"
	"github.com/celestiaorg/celestia-node/params"
)

// settings store values that can be augmented or changed for Node with Options.
type settings struct {
	cfg  *Config
	opts []fx.Option
}

// Option for Node's Config.
type Option func(*settings)

// WithNetwork specifies the Network to which the Node should connect to.
// WARNING: Use this option with caution and never run the Node with different networks over the same persisted Store.
func WithNetwork(net params.Network) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(net))
	}
}

// WithP2PKey sets custom Ed25519 private key for p2p networking.
func WithP2PKey(key crypto.PrivKey) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(fx.Annotate(key, fx.As(new(crypto.PrivKey)))))
	}
}

// WithP2PKeyStr sets custom hex encoded Ed25519 private key for p2p networking.
func WithP2PKeyStr(key string) Option {
	return func(sets *settings) {
		decKey, err := hex.DecodeString(key)
		if err != nil {
			sets.opts = append(sets.opts, fx.Error(err))
			return
		}

		key, err := crypto.UnmarshalEd25519PrivateKey(decKey)
		if err != nil {
			sets.opts = append(sets.opts, fx.Error(err))
			return
		}

		sets.opts = append(sets.opts, fx.Replace(fx.Annotate(key, fx.As(new(crypto.PrivKey)))))
	}

}

// WithHost sets custom Host's data for p2p networking.
func WithHost(hst host.Host) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(fx.Annotate(hst, fx.As(new(p2p.HostBase)))))
	}
}

// WithCoreClient sets custom client for core process
func WithCoreClient(client core.Client) Option {
	return func(sets *settings) {
		sets.opts = append(sets.opts, fx.Replace(fx.Annotate(client, fx.As(new(core.Client)))))
	}
}
