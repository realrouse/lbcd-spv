#!/usr/bin/env bash
# Regtest: one full node (any CF-capable lbcd binary) funds an SPV wallet.
#   ./scripts/fetch-lbcd.sh   # or export LBCD_BIN=/path/to/lbcd
#   ./scripts/harness-spv.sh

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export PATH="$ROOT/bin:$PATH"

if [[ -z "${LBCD_BIN:-}" ]]; then
  if [[ -x "$ROOT/bin/lbcd" ]]; then
    LBCD_BIN="$ROOT/bin/lbcd"
  elif command -v lbcd >/dev/null; then
    LBCD_BIN="$(command -v lbcd)"
  else
    echo "No lbcd binary. Run ./scripts/fetch-lbcd.sh or set LBCD_BIN." >&2
    exit 1
  fi
fi
if [[ -z "${LBCCTL_BIN:-}" ]]; then
  if [[ -x "$ROOT/bin/lbcctl" ]]; then
    LBCCTL_BIN="$ROOT/bin/lbcctl"
  elif command -v lbcctl >/dev/null; then
    LBCCTL_BIN="$(command -v lbcctl)"
  else
    LBCCTL_BIN=""
  fi
fi
if [[ -z "${LBCWALLET_BIN:-}" ]]; then
  if [[ -x "$ROOT/bin/lbcwallet" ]]; then
    LBCWALLET_BIN="$ROOT/bin/lbcwallet"
  else
    echo "Building lbcwallet..."
    (cd "$ROOT/lbcwallet" && go build -o "$ROOT/bin/lbcwallet" .)
    LBCWALLET_BIN="$ROOT/bin/lbcwallet"
  fi
fi

echo "Using LBCD_BIN=$LBCD_BIN"
echo "Using LBCWALLET_BIN=$LBCWALLET_BIN"

WORKDIR="${HARNESS_DIR:-$ROOT/.harness-spv}"
rm -rf "$WORKDIR"
mkdir -p "$WORKDIR"/{alpha,spv}

RPC_USER=user
RPC_PASS=pass
NODE_P2P=127.0.0.1:29246
NODE_RPC=127.0.0.1:29245
ALPHA_WALLET_RPC=127.0.0.1:29244
SPV_WALLET_RPC=127.0.0.1:29254
NODE_PID=""
ALPHA_WALLET_PID=""
SPV_PID=""

ctl() {
  if [[ -n "$LBCCTL_BIN" ]]; then
    "$LBCCTL_BIN" --regtest --rpcuser="$RPC_USER" --rpcpass="$RPC_PASS" --rpcserver="$NODE_RPC" --notls "$@"
  else
    curl -s --user "$RPC_USER:$RPC_PASS" \
      --data-binary "{\"jsonrpc\":\"1.0\",\"id\":\"1\",\"method\":\"$1\",\"params\":$(python3 -c 'import json,sys; print(json.dumps(sys.argv[1:]))' "${@:2}")}" \
      -H 'content-type: text/plain;' \
      "http://${NODE_RPC}/"
  fi
}

wallet_rpc() {
  local url="$1" method="$2"; shift 2
  python3 - "$url" "$RPC_USER" "$RPC_PASS" "$method" "$@" <<'PY'
import json,sys,urllib.request,base64
url,user,pw,method=sys.argv[1],sys.argv[2],sys.argv[3],sys.argv[4]
params=sys.argv[5:]
# coerce numbers
out=[]
for p in params:
    try:
        if "." in p: out.append(float(p))
        else: out.append(int(p))
    except ValueError:
        out.append(p)
body=json.dumps({"jsonrpc":"1.0","id":"1","method":method,"params":out}).encode()
req=urllib.request.Request("http://"+url+"/", data=body, headers={"Content-Type":"text/plain"})
token=base64.b64encode(f"{user}:{pw}".encode()).decode()
req.add_header("Authorization","Basic "+token)
resp=json.load(urllib.request.urlopen(req, timeout=30))
if resp.get("error"):
    raise SystemExit(f"{method} error: {resp['error']}")
r=resp.get("result")
if isinstance(r,(dict,list)):
    print(json.dumps(r))
else:
    print(r)
PY
}

echo "Starting full node..."
"$LBCD_BIN" --regtest --notls --txindex \
  --datadir="$WORKDIR/alpha/data" --logdir="$WORKDIR/alpha/logs" \
  --listen="$NODE_P2P" --rpclisten="$NODE_RPC" \
  --rpcuser="$RPC_USER" --rpcpass="$RPC_PASS" \
  >"$WORKDIR/alpha/node.log" 2>&1 &
NODE_PID=$!

cleanup() {
  kill ${SPV_PID:-} ${ALPHA_WALLET_PID:-} ${NODE_PID:-} 2>/dev/null || true
}
trap cleanup EXIT

for i in $(seq 1 30); do
  if ctl getblockcount >/dev/null 2>&1; then break; fi
  sleep 0.3
done

echo "Starting alpha (full-node) wallet for mining..."
"$LBCWALLET_BIN" --createtemp --regtest --noservertls --noclienttls \
  --appdata="$WORKDIR/alpha/wallet" \
  --rpcuser="$RPC_USER" --rpcpass="$RPC_PASS" \
  --rpclisten="$ALPHA_WALLET_RPC" --rpcconnect="$NODE_RPC" \
  >"$WORKDIR/alpha/wallet.log" 2>&1 &
ALPHA_WALLET_PID=$!

for i in $(seq 1 40); do
  if wallet_rpc "$ALPHA_WALLET_RPC" getnewaddress >/dev/null 2>&1; then break; fi
  sleep 0.3
done
ADDR="$(wallet_rpc "$ALPHA_WALLET_RPC" getnewaddress)"
echo "Mining address $ADDR"
ctl generatetoaddress 120 "$ADDR" >/dev/null
echo "Node height $(ctl getblockcount)"
echo "Waiting for alpha wallet to see mature coinbase..."
for i in $(seq 1 60); do
  ABAL="$(wallet_rpc "$ALPHA_WALLET_RPC" getbalance || echo 0)"
  python3 -c "import sys; sys.exit(0 if float('$ABAL')>=10 else 1)" && break
  sleep 0.5
done
echo "Alpha getbalance=$ABAL"

echo "Starting SPV wallet (createtemp + --spv, peers only)..."
"$LBCWALLET_BIN" --createtemp --regtest --spv --noservertls --noclienttls \
  --appdata="$WORKDIR/spv/wallet" \
  --connect="$NODE_P2P" --nodnsseed \
  --rpcuser="$RPC_USER" --rpcpass="$RPC_PASS" \
  --rpclisten="$SPV_WALLET_RPC" \
  >"$WORKDIR/spv/wallet.log" 2>&1 &
SPV_PID=$!

echo "Waiting for SPV wallet RPC..."
for i in $(seq 1 60); do
  if wallet_rpc "$SPV_WALLET_RPC" getnewaddress >/dev/null 2>&1; then break; fi
  sleep 0.5
done
SPV_ADDR="$(wallet_rpc "$SPV_WALLET_RPC" getnewaddress)"
echo "SPV receive address $SPV_ADDR"
echo "Waiting for SPV header/filter sync..."
sleep 5

echo "Sending 10 LBC from alpha to SPV..."
wallet_rpc "$ALPHA_WALLET_RPC" sendtoaddress "$SPV_ADDR" 10 >/dev/null
ctl generatetoaddress 1 "$ADDR" >/dev/null

echo "Waiting for SPV balance..."
BAL=0
for i in $(seq 1 40); do
  BAL="$(wallet_rpc "$SPV_WALLET_RPC" getbalance || echo 0)"
  python3 -c "import sys; sys.exit(0 if float('$BAL')>=10 else 1)" && break
  sleep 1
done
echo "SPV getbalance=$BAL"
python3 -c "import sys; sys.exit(0 if float('$BAL')>=10 else 1)"

echo "Sending 1 LBC from SPV back to alpha..."
wallet_rpc "$SPV_WALLET_RPC" sendtoaddress "$ADDR" 1 >/dev/null
ctl generatetoaddress 1 "$ADDR" >/dev/null
sleep 2
echo "SPV getbalance=$(wallet_rpc "$SPV_WALLET_RPC" getbalance)"

echo "PASS: SPV wallet received and spent on regtest against $LBCD_BIN"
