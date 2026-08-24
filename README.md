# LBC SPV wallet

## In plain language

This is a **light LBRY Credits (LBC) wallet**. It can receive and send LBC **without downloading the whole blockchain**.

Think of two kinds of software:

| | Full node (`lbcd`) | This SPV wallet (`lbcwallet --spv`) |
|---|---|---|
| What it stores | The entire chain (blocks, claims, indexes) | Only block headers, compact filters, and *your* coins |
| Disk | Official docs: **≥ 100 GB** (grows slowly) | A few hundred MB to a few GB |
| RAM | Official docs: **≥ 8 GB** | A small process |
| What it can see | Every transaction on the network | Only payments that belong to this wallet |
| Who needs it | DEX **server**, explorers, people serving light wallets | Traders, market makers, phones, small VPS wallets |

This repo is **only the light wallet**. It still needs **someone else** running a full node that serves compact filters. That node *is* the SPV “server.” We did not build a separate server binary.

Who runs what:

- **Neofutur / a dedicated box** → `lbcd` (full node, filters on). This is the server light wallets use.
- **Traders / Bison Wallet / this repo** → `lbcwallet --spv`. Thin client. Points at that node with `--connect`.

Many public LBC peers are still old `lbrycrd` and **do not** serve compact filters. Until the public network has plenty of `lbcd` nodes, **someone must keep at least one filter-serving node online** or SPV wallets cannot sync.

It does **not** replace a full node for a dcrdex **server**. The matching engine must watch *everyone’s* swap contracts, not just one wallet.

Any full node that speaks LBC P2P and serves compact filters works — lbryio builds, a future LBRY Foundation build, or a custom daemon. `--connect=host:port`. The wallet does not care which GitHub org compiled the node.

---

A Neutrino (BIP 157/158) light client and `lbcwallet --spv` daemon for LBRY Credits.

This wallet syncs **headers and compact filters over P2P**. It does not need a local full node on the same machine.

The dcrdex **server** still needs a full node. This software is for a later Bison Wallet / dcrdex **client** integration.

### If you are renting a dedicated server (copy this)

**Yes — you are the SPV server.** That machine should run **`lbcd`**, not this light wallet.

A light wallet does not work by itself. It downloads headers and compact filters from a full node. Your **~500 GB SSD / 16–64 GB RAM** box is that full node. Compact filters are **on by default** in lbcd (do not pass `--nocfilters`).

Official lbcd asks for at least 8 GB RAM and 100 GB disk. 16 GB RAM and 500 GB NVMe is comfortable for chain growth, indexes, serving light wallets, and later a dcrdex server on the same host.

Then:

1. Keep P2P port **9246** reachable (this is what SPV clients connect to). Firewall it to people you trust if you do not want the whole internet.
2. Do **not** expose RPC (9245) to the internet. RPC stays on localhost (dcrdex server / you).
3. Light wallets connect with `--connect=your.server:9246 --nodnsseed`.
4. One `lbcd` can serve many SPV wallets at once. Same box can later run the dcrdex **server** against local RPC.

Binaries today: [lbryio/lbcd releases](https://github.com/lbryio/lbcd/releases). [LBRYFoundation/lbcd](https://github.com/LBRYFoundation/lbcd) is the likely future maintainer but does not publish binaries yet — same protocol either way.

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
