# LBC SPV wallet

A Neutrino (BIP 157/158) light client and `lbcwallet --spv` daemon for LBRY Credits.

This wallet syncs **headers and compact filters over P2P**. It does not need a local full node, and it does not care which GitHub org built the peer you connect to. Any daemon that speaks the LBC wire protocol and serves compact filters (`SFNodeCF`) works.

The dcrdex **server** still needs a full node. This software is for a later Bison Wallet / dcrdex **client** integration.

## What you get

| Tool | Role |
|------|------|
| `neutrino` library | Header + compact-filter sync |
| `cmd/neutrino` | Sync headers/filters and print the tip |
| `cmd/cfprobe` | Ask a peer (or DNS seeds) whether it serves compact filters |
| `lbcwallet --spv` | HD wallet JSON-RPC on port 9244, no node RPC |

## Full node binaries (default)

Harness scripts download **release binaries** from [lbryio/lbcd](https://github.com/lbryio/lbcd/releases) because that is the published build today. [LBRYFoundation/lbcd](https://github.com/LBRYFoundation/lbcd) is expected to take over maintenance later but does not ship binaries yet.

Override any time:

```bash
# Use a Foundation build, a custom fork, or a binary you compiled yourself
export LBCD_BIN=/path/to/your/lbcd
export LBCD_RELEASE_URL=https://example.invalid/lbcd.tar.gz   # optional, for fetch-lbcd.sh
```

The wallet never clones a GitHub repo and never hard-codes an organization name as “the” daemon.

## Quick start (regtest)

```bash
./scripts/fetch-lbcd.sh          # or set LBCD_BIN
./scripts/harness-spv.sh
```

## SPV wallet against your own node

```bash
lbcwallet --create -p 'a-real-passphrase'
lbcwallet --spv --connect=127.0.0.1:9246 --nodnsseed \
  --rpcuser=user --rpcpass=pass
```

`--connect` is any host:port. `--nodnsseed` disables public DNS seeds so you only talk to the peers you named.

## Layout

```
neutrino/     BIP157/158 light client (LBC wire, LBRY PoW, 112-byte headers)
lbcwallet/    HD wallet daemon with --spv
cmd/          neutrino + cfprobe CLIs
scripts/      fetch-lbcd.sh, harness-spv.sh
docs/         architecture and RPC notes
```

## Networks

| Network | P2P | Wallet RPC |
|---------|-----|------------|
| mainnet | 9246 | 9244 |
| testnet | 19246 | 19244 |
| regtest | 29246 | 29244 |
