package headerfs

import (
	"bytes"
	"testing"

	"github.com/lbryio/lbcd/chaincfg"
	"github.com/lbryio/lbcd/wire"
)

func TestLBCHeaderIs112Bytes(t *testing.T) {
	if BlockHeaderSize != 112 {
		t.Fatalf("BlockHeaderSize=%d, want 112", BlockHeaderSize)
	}
	if wire.MaxBlockHeaderPayload != BlockHeaderSize {
		t.Fatalf("wire.MaxBlockHeaderPayload=%d, want %d",
			wire.MaxBlockHeaderPayload, BlockHeaderSize)
	}

	hdr := chaincfg.MainNetParams.GenesisBlock.Header
	var buf bytes.Buffer
	if err := hdr.Serialize(&buf); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != BlockHeaderSize {
		t.Fatalf("serialized genesis header is %d bytes, want %d", buf.Len(), BlockHeaderSize)
	}

	var round wire.BlockHeader
	if err := round.Deserialize(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal(err)
	}
	if round.BlockHash() != hdr.BlockHash() {
		t.Fatalf("round-trip hash %v != %v", round.BlockHash(), hdr.BlockHash())
	}
	if round.ClaimTrie != hdr.ClaimTrie {
		t.Fatalf("ClaimTrie was dropped: got %v want %v", round.ClaimTrie, hdr.ClaimTrie)
	}
}
