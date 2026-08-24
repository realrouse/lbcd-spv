# Upstream

This tree started from [lightninglabs/neutrino v0.12.1](https://github.com/lightninglabs/neutrino/tree/v0.12.1).

v0.12.1 is the last Neutrino release whose blockchain API matches `lbryio/lbcd` (CheckProofOfWork, CheckBlockSanity, no btcd v2 modules, no HeaderCtx interface). Newer Neutrino requires `CheckBlockHeaderContext` which lbcd does not export.

LBC-specific changes on top of that tag:

- Import `github.com/lbryio/lbcd` and `github.com/lbryio/lbcutil` instead of btcsuite.
- Store 112-byte ClaimTrie headers (`headerfs.BlockHeaderSize`).
- Use lbcd's LBRY PoW check (`BlockPoWHash`) via `blockchain.CheckProofOfWork`.
- Difficulty clamp copied from lbcd (retarget every block; ±1/8 and +1/2 timespan).
- Require `SFNodeNetwork|SFNodeCF` so a custom daemon does not have to advertise witness.
- `--connect` / DNS-seed override so the peer is not tied to a GitHub org.
- Drop Bitcoin filter-header checkpoints (wrong chain).
