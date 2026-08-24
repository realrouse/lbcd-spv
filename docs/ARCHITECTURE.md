# Architecture

## Runtime vs compile-time

The **running** SPV wallet speaks LBC P2P plus compact filters. Peers are chosen with `--connect`, `--addpeer`, and `--dnsseed` / `--nodnsseed`. No GitHub organization is baked into the binary.

**Compile-time** consensus types (112-byte ClaimTrie headers, LBRY PoW, GCS filters) come from `github.com/lbryio/lbcd` and `github.com/lbryio/lbcutil` because those modules exist today. When Foundation publishes a module, retarget with:

```
replace github.com/lbryio/lbcd => github.com/LBRYFoundation/lbcd <version>
```

## Why Neutrino

dcrdex already uses Neutrino for native BTC/LTC/BCH wallets. LBRY hubs (Electrum-style) leak addresses and are not the path dcrdex embeds.

`lbcwallet` already had `chain.Interface` “RPC or SPV library” and used `getcfilter` when talking to node RPC. It never shipped a P2P Neutrino backend. This repo adds that backend.

## LBC deltas from Bitcoin Neutrino

1. **112-byte headers** — Version, PrevBlock, MerkleRoot, **ClaimTrie**, Timestamp, Bits, Nonce. Header files must not assume 80 bytes.
2. **Split hashes** — `BlockHash()` is SHA256d (identity, filter key, prev-block). `BlockPoWHash()` is LBRY’s SHA256d→SHA512→RIPEMD160 pair→SHA256d (difficulty). `lbcd/blockchain.CheckProofOfWork` already uses the PoW hash.
3. **Retarget every block** — `TargetTimespan == TargetTimePerBlock` (150s), so Neutrino’s `blocksPerRetarget` is 1. Clamp matches lbcd: `timespan ± 1/8 and +1/2`, not Bitcoin’s ×4 / ÷4.
4. **Compact filters** — BIP158 basic filters from `lbcutil/gcs`, keyed by LBC `BlockHash()`. Claim scripts and P2SH swap contracts are ordinary scriptPubKeys.
5. **CF peers** — Many public nodes may still be lbrycrd without BIP157. Pin `--connect` at an operator node that serves filters. `cmd/cfprobe` surveys that.

## Later dcrdex client

Native `walletTypeSPV` in `client/asset/lbc` (like `client/asset/ltc/spv.go`) is a follow-up. Until then, `lbcwallet --spv` JSON-RPC is the integration surface. See `docs/RPC-SPV.md`.

The dcrdex **server** asset backend stays on a full node.
