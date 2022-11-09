package swamp

import (
	"time"

	tn "github.com/tendermint/tendermint/config"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/celestiaorg/celestia-app/testutil/testnode"
)

// Components struct represents a set of pre-requisite attributes from the test scenario
type Components struct {
	CoreCfg         *tn.Config
	ConsensusParams *tmproto.ConsensusParams
	SupressLogs     bool
}

// DefaultComponents creates a celestia-app instance with a block time of around
// 100ms
func DefaultComponents() *Components {
	tnCfg := tn.TestConfig()
	tnCfg.Consensus.TimeoutCommit = 100 * time.Millisecond
	return &Components{
		CoreCfg:         tnCfg,
		ConsensusParams: testnode.DefaultParams(),
		SupressLogs:     true,
	}
}

// Option for modifying Swamp's Config.
type Option func(*Components)

// WithBlockTime sets a custom interval for block creation.
func WithBlockTime(t time.Duration) Option {
	return func(c *Components) {
		// for empty block
		c.CoreCfg.Consensus.CreateEmptyBlocksInterval = t
		// for filled block
		c.CoreCfg.Consensus.TimeoutCommit = t
		c.CoreCfg.Consensus.SkipTimeoutCommit = false
	}
}
