# Draft PR — not opened yet

**Target:** https://github.com/LBRYFoundation/lbcwallet  
(fork of `lbryio/lbcwallet`; Go module path is still `github.com/lbryio/lbcwallet`)

**Do not file this until** the Neutrino module is importable as
`github.com/realrouse/lbcd-spv/neutrino` (today the `go.mod` still says
`github.com/lbc-spv/neutrino`, which `go get` cannot fetch).

---

## Title

```
chain: add NeutrinoClient so the wallet library can sync over compact filters
```

## Body (paste into GitHub)

```markdown
## Why

`chain.Interface` already says the backend can be “an RPC chain server, **or an
SPV library**, as long as we write a driver for it.” Only the RPC driver was
ever written. README still says lbcwallet is not an SPV client.

Bison Wallet / dcrdex native BTC SPV does **not** fork `btcwallet`. It imports
upstream as a library:

- `github.com/btcsuite/btcwallet/wallet` (Loader, addresses, tx store)
- `github.com/btcsuite/btcwallet/chain` (`NeutrinoClient` implements `chain.Interface`)
- `github.com/lightninglabs/neutrino` (`ChainService`)

We want the same shape for LBC, so dcrdex can add `client/asset/lbc` native SPV
without shipping a forked `lbcwallet` binary.

## What this PR does

Adds the missing SPV **driver** next to `RPCClient`:

- `chain.NeutrinoClient` implements `chain.Interface` on top of
  `github.com/realrouse/lbcd-spv/neutrino` (LBC port of Neutrino: 112-byte
  ClaimTrie headers, LBRY PoW, per-block difficulty).
- `BackEnds()` includes `"neutrino"`.
- Default RPC mode is unchanged. No JSON-RPC passthrough to a full node in
  SPV mode (there is no node RPC).

This is a **library** change. Callers (dcrdex, or a later `--spv` CLI patch)
construct `neutrino.ChainService` and wrap it with `chain.NewNeutrinoClient`.

## What this PR does not do

- It does **not** add `lbcwallet --spv` CLI flags (optional follow-up if you
  want a standalone daemon).
- It does **not** replace `lbcd`. Compact-filter **full nodes** still must run
  `lbcd` with filters left on and P2P `:9246` reachable. Light clients connect
  to that node.
- It does **not** put Neutrino itself in this repository. Neutrino for LBC is
  a separate module (same split as `btcwallet` vs `lightninglabs/neutrino`).

## How dcrdex would use it (Path A)

```go
db, _ := walletdb.Create("bdb", neutrinoDBPath, true, timeout)
svc, _ := neutrino.NewChainService(neutrino.Config{
    DataDir: dir, Database: db, ChainParams: *params,
    ConnectPeers: []string{"full-node:9246"}, DisableDNSSeed: true,
})
ncl := chain.NewNeutrinoClient(params, svc)
loader := wallet.NewLoader(params, dir, true, timeout, 250)
w, _ := loader.OpenExistingWallet(...)
w.SynchronizeRPC(ncl)
```

Same pattern as `decred/dcrdex` `client/asset/btc/spv.go`.

## Testing

Regtest (full `lbcd` + this client): create wallet, sync headers/filters,
receive a payment, spend it back. Compact filters must be enabled on the
`lbcd` under test (default).

## Dependency

```
require github.com/realrouse/lbcd-spv/neutrino v0.1.0
```

Happy to point this at a Foundation-hosted fork of Neutrino later if you
prefer not to depend on `realrouse/lbcd-spv`.
```

---

## Files the PR would touch

| File | Change |
|------|--------|
| `chain/neutrino_client.go` | **New.** Copy of `lbcwallet/chain/neutrino_client.go` from `realrouse/lbcd-spv`, imports switched to the published Neutrino module. |
| `chain/neutrino.go` | Keep existing `buildFilterBlocksWatchList` (already used by RPC `FilterBlocks`). |
| `chain/interface.go` | Add `"neutrino"` to `BackEnds()`. |
| `go.mod` / `go.sum` | Require `github.com/realrouse/lbcd-spv/neutrino`. |
| `README.md` | Replace “is not an SPV client” with: library can use Neutrino; default binary is still RPC-to-`lbcd`; a filter-serving `lbcd` is still required. |

**Not in this PR:** `spv.go`, `--spv` / `--connect` / `--nodnsseed` flags, our copy of the whole wallet tree.

## Optional second PR (not needed for dcrdex)

CLI `lbcwallet --spv --connect=host:9246 --nodnsseed` so people can run a
standalone light daemon. dcrdex Path A does not use that binary.

## Why bother Foundation at all?

You can skip this PR and put `NeutrinoClient` inside dcrdex. That still works.

A Foundation PR is worth it because:

1. `chain.Interface` already reserved this driver; BTC put `NeutrinoClient` in
   `btcwallet/chain`, not in dcrdex.
2. dcrdex then depends on **stock** `lbcwallet` + Neutrino — no wallet fork.
3. Other LBC apps get SPV without copying 400 lines.

It is **not** a blocker for starting `client/asset/lbc` SPV.
