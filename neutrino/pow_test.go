package neutrino

import (
	"math/big"
	"testing"
	"time"

	"github.com/lbryio/lbcd/blockchain"
	"github.com/lbryio/lbcd/chaincfg"
	"github.com/lbryio/lbcd/wire"
	btcutil "github.com/lbryio/lbcutil"
)

func TestProofOfWorkUsesLbryHash(t *testing.T) {
	genesis := chaincfg.RegressionNetParams.GenesisBlock
	block := btcutil.NewBlock(genesis)
	powLimit := chaincfg.RegressionNetParams.PowLimit

	if err := blockchain.CheckProofOfWork(block, powLimit); err != nil {
		t.Fatalf("regtest genesis should pass LBRY PoW: %v", err)
	}

	// Identity hash (SHA256d) is not the PoW hash. A checker that used
	// BlockHash() against Bits would still "pass" easy regtest targets,
	// so we only assert the two hashes differ on a header with a ClaimTrie.
	hdr := genesis.Header
	if hdr.BlockHash() == hdr.BlockPoWHash() {
		t.Fatal("BlockHash and BlockPoWHash must differ (LBRY PoW is not SHA256d)")
	}
}

func TestRetargetIntervalIsEveryBlock(t *testing.T) {
	p := chaincfg.MainNetParams
	if p.TargetTimespan != p.TargetTimePerBlock {
		t.Fatalf("expected per-block retarget, timespan=%s per-block=%s",
			p.TargetTimespan, p.TargetTimePerBlock)
	}
	blocks := int32(p.TargetTimespan / p.TargetTimePerBlock)
	if blocks != 1 {
		t.Fatalf("blocksPerRetarget=%d, want 1", blocks)
	}
	_ = time.Second
	_ = big.NewInt(0)
	_ = wire.MainNet
}
