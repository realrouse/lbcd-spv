# How to run the SPV “server” (a full `lbcd` node)

There is no extra SPV-server program. **Install `lbcd`, leave compact filters on, open P2P, keep RPC private.** Light wallets (`lbcwallet --spv`) then connect to you on port **9246**.

Hardware that is enough (official `lbcd` docs): **≥ 8 GB RAM, ≥ 100 GB disk**, fast SSD. A **500 GB NVMe + 16–64 GB RAM** box is comfortable.

## 1. Get the binary

Today: [lbryio/lbcd releases](https://github.com/lbryio/lbcd/releases) (e.g. `lbcd_*_linux_amd64.tar.gz`).

Unpack `lbcd` and `lbcctl` into `/usr/local/bin`.

(When [LBRYFoundation/lbcd](https://github.com/LBRYFoundation/lbcd) publishes builds, use those instead. Same ports, same protocol.)

## 2. Create a user and config

```bash
sudo useradd --system --home /var/lib/lbcd --create-home --shell /usr/sbin/nologin lbcd
sudo mkdir -p /var/lib/lbcd
sudo chown lbcd:lbcd /var/lib/lbcd
```

`/var/lib/lbcd/lbcd.conf`:

```ini
; Compact filters stay ON. Do not set nocfilters=1.
txindex=1
addrindex=1

listen=0.0.0.0:9246
rpclisten=127.0.0.1:9245
rpcuser=CHANGE_ME
rpcpass=CHANGE_ME_TO_A_LONG_RANDOM_SECRET

maxpeers=125
```

- **`listen=0.0.0.0:9246`** — P2P. This is what SPV wallets use. Must be reachable from the internet (or from the people who will run light wallets).
- **`rpclisten=127.0.0.1:9245`** — JSON-RPC. Local only. Never public. You (and later dcrdex) use this.
- **`txindex=1`** — not required for SPV clients, but needed if this same node later backs a dcrdex **server**. Cheaper to turn on before first sync.
- Do **not** add `nocfilters=1`. Filters are how SPV works.

## 3. Firewall

```bash
# P2P for the whole net (or restrict to known IPs)
sudo ufw allow 9246/tcp comment 'LBC P2P / SPV'
# RPC stays closed
sudo ufw deny 9245/tcp
```

If the machine is behind NAT, forward **TCP 9246** to it.

## 4. Run it (systemd)

`/etc/systemd/system/lbcd.service`:

```ini
[Unit]
Description=LBRY Credits full node (compact-filter / SPV peer)
After=network-online.target
Wants=network-online.target

[Service]
User=lbcd
Group=lbcd
ExecStart=/usr/local/bin/lbcd --appdata=/var/lib/lbcd
Restart=always
RestartSec=10
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now lbcd
sudo journalctl -u lbcd -f
```

Wait until it is at chain tip (`getblockcount` stops jumping and matches an explorer). First sync can take hours to days. Optional: load a snapshot from https://snapshots.lbry.com/blockchain/ then start `lbcd` so it only has to catch up.

## 5. Check that it will serve SPV clients

On the server:

```bash
lbcctl --rpcuser=CHANGE_ME --rpcpass=CHANGE_ME getpeerinfo
lbcctl --rpcuser=CHANGE_ME --rpcpass=CHANGE_ME getcfilter \
  "$(lbcctl --rpcuser=CHANGE_ME --rpcpass=CHANGE_ME getbestblockhash)" 0
```

`getcfilter` must return filter bytes, not an error. If it errors with no CF index, you started with `nocfilters` or the index is still building.

From a laptop:

```bash
# from this repo
go run ./neutrino/cmd/cfprobe --connect=YOUR.PUBLIC.IP:9246
```

You want to see `SFNodeCF` for that peer.

## 6. Tell light-wallet users

```
lbcwallet --spv --connect=YOUR.PUBLIC.IP:9246 --nodnsseed
```

One `lbcd` can serve many SPV wallets. You do **not** run `lbcwallet --spv` on this dedicated server unless you also want a wallet there.

## Do / don’t

| Do | Don’t |
|----|--------|
| Leave compact filters on | `--nocfilters` |
| Open **9246** | Open **9245** to the internet |
| Strong `rpcuser` / `rpcpass` | Default or public RPC |
| Keep the process running 24/7 | Stop it when people need to sync wallets |
| `--txindex` if dcrdex server comes later | Assume old `lbrycrd` is enough (it usually has no BIP157 filters) |
