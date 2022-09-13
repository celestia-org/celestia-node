package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/celestiaorg/celestia-node/node"
	"github.com/celestiaorg/celestia-node/params"
)

var (
	nodeStoreFlag   = "node.store"
	nodeConfigFlag  = "node.config"
	nodeNetworkFlag = "node.network"
)

// NodeFlags gives a set of hardcoded Node package flags.
func NodeFlags() *flag.FlagSet {
	flags := &flag.FlagSet{}

	flags.String(
		nodeStoreFlag,
		"",
		"The path to root/home directory of your Celestia Node Store",
	)
	flags.String(
		nodeConfigFlag,
		"",
		"Path to a customized node config TOML file",
	)
	flags.String(
		nodeNetworkFlag,
		"",
		"The name of the network to connect to, e.g. "+params.ListProvidedNetworks(),
	)

	return flags
}

// ParseNodeFlags parses Node flags from the given cmd and applies values to Env.
func ParseNodeFlags(ctx context.Context, cmd *cobra.Command) (context.Context, error) {
	network := cmd.Flag(nodeNetworkFlag).Value.String()
	if network != "" {
		err := params.SetDefaultNetwork(params.Network(network))
		if err != nil {
			return ctx, err
		}
	} else {
		network = string(params.DefaultNetwork())
	}

	store := cmd.Flag(nodeStoreFlag).Value.String()
	if store == "" {
		tp := NodeType(ctx)
		store = fmt.Sprintf("~/.celestia-%s-%s", strings.ToLower(tp.String()), strings.ToLower(network))
	}
	ctx = WithStorePath(ctx, store)

	nodeConfig := cmd.Flag(nodeConfigFlag).Value.String()
	if nodeConfig != "" {
		cfg, err := node.LoadConfig(nodeConfig)
		if err != nil {
			return ctx, fmt.Errorf("cmd: while parsing '%s': %w", nodeConfigFlag, err)
		}

		ctx = WithNodeOptions(ctx, node.WithConfig(cfg))
	}

	return ctx, nil
}
