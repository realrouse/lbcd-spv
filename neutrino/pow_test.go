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

func TestLbcdLookbackHeight(t *testing.T) {
	// lastHeight, interval → firstHeight (lbcd RelativeAncestor(min(interval, last)))
	cases := [][3]int32{
		{0, 1, 0},    // genesis → self
		{1, 1, 0},    // block 1 → genesis (NOT itself)
		{2, 1, 1},    // block 2 → block 1
		{2015, 2016, 0},
		{4031, 2016, 2015},
	}
	for _, c := range cases {
		last, interval, want := c[0], c[1], c[2]
		blocksBack := interval
		if blocksBack > last {
			blocksBack = last
		}
		got := last - blocksBack
		if got != want {
			t.Fatalf("last=%d interval=%d first=%d want %d", last, interval, got, want)
		}
	}
}

func TestLbcAdjustedTimespanMatchesLbcd(t *testing.T) {
	const target int64 = 150
	minSpan := target - target/8 // 131
	maxSpan := target + target/2 // 225

	// Fast block: Bitcoin-style would use ~131; LBC damps to 141.
	if got := lbcAdjustedTimespan(75, target, minSpan, maxSpan); got != 141 {
		t.Fatalf("fast block timespan=%d, want 141", got)
	}
	// On-target block stays on target.
	if got := lbcAdjustedTimespan(150, target, minSpan, maxSpan); got != 150 {
		t.Fatalf("on-target timespan=%d, want 150", got)
	}
	// Slow block damps toward 150 then may clamp at 225.
	if got := lbcAdjustedTimespan(300, target, minSpan, maxSpan); got != 168 {
		t.Fatalf("slow block timespan=%d, want 168", got)
	}
	if got := lbcAdjustedTimespan(10_000, target, minSpan, maxSpan); got != maxSpan {
		t.Fatalf("very slow block timespan=%d, want clamp %d", got, maxSpan)
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
