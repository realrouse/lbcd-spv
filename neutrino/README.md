# Neutrino for LBRY Credits

A BIP 157/158 light client for LBC, ported from lightninglabs/neutrino v0.12.1.

See `UPSTREAM.md` for the tag we started from and the LBC-specific changes (112-byte ClaimTrie headers, LBRY proof-of-work, per-block difficulty).

```bash
go run ./cmd/neutrino --network=regtest --connect=127.0.0.1:29246 --nodnsseed
go run ./cmd/cfprobe --connect=127.0.0.1:9246
```

`--connect` is any compact-filter-capable LBC daemon. It does not have to come from a particular GitHub organization.
