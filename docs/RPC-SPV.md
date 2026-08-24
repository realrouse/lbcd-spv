# JSON-RPC in SPV mode

`lbcwallet --spv` keeps the wallet JSON-RPC on the usual ports. There is **no** passthrough to a full node, because there is no node RPC.

## Wallet methods (same as full-node mode)

These hit the wallet database and work in SPV:

`getbalance`, `getnewaddress`, `getrawchangeaddress`, `listunspent`, `lockunspent`, `listlockunspent`, `signrawtransaction`, `sendtoaddress`, `walletpassphrase`, `walletlock`, `dumpprivkey`, `gettransaction`, `validateaddress`, `listsinceblock`

## Neutrino-backed (via chain.Interface, not HTTP passthrough)

The wallet already uses the chain backend for send/rescan. Extra node-style RPCs are **not** proxied. Until native dcrdex SPV lands, callers that need these should use wallet methods or wait for dedicated handlers:

| Method | SPV status |
|--------|------------|
| getblockcount / getbestblockhash / getblockhash / getblockheader / getblock | unsupported over HTTP (available in-process on Neutrino) |
| sendrawtransaction | used internally when the wallet broadcasts |
| getblockchaininfo / getnetworkinfo | unsupported over HTTP |
| estimatesmartfee | unsupported; use a fallback fee |
| gettxout / getrawtransaction | wallet db only; unknown txs return not found |
| getrawmempool | unsupported |

dcrdex `client/asset/btc` RPC clone calls several of the node methods. That is why the **later** Bison integration should be native `walletTypeSPV` (in-process Neutrino), not HTTP passthrough.

## Later dcrdex client

Model on `client/asset/ltc/spv.go`: seeded wallet type `"SPV"`, type-translate `lbryio/lbcd` ↔ `btcsuite/btcd` where the clone layer needs it, `OpenSPVWallet` using this Neutrino + lbcwallet stack.
