// Command neutrino syncs LBC headers and compact-filter headers over P2P
// and prints the chain tip. It does not require a local full node.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/btcsuite/btcwallet/walletdb"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/lbryio/lbcd/chaincfg"
	"github.com/lbc-spv/neutrino"
)

func main() {
	datadir := flag.String("datadir", filepath.Join(".", "neutrino-data"), "header and filter database directory")
	network := flag.String("network", "mainnet", "mainnet, testnet, or regtest")
	connect := flag.String("connect", "", "comma-separated host:port peers (only these if set)")
	addpeer := flag.String("addpeer", "", "comma-separated extra persistent peers")
	nodnsseed := flag.Bool("nodnsseed", false, "do not query DNS seeds")
	dnsseed := flag.String("dnsseed", "", "comma-separated DNS seeds (replaces defaults if set)")
	flag.Parse()

	params := chaincfg.MainNetParams
	switch strings.ToLower(*network) {
	case "mainnet", "main":
		params = chaincfg.MainNetParams
	case "testnet", "testnet3":
		params = chaincfg.TestNet3Params
	case "regtest", "simnet":
		params = chaincfg.RegressionNetParams
	default:
		fmt.Fprintf(os.Stderr, "unknown network %q\n", *network)
		os.Exit(2)
	}

	if err := os.MkdirAll(*datadir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dbPath := filepath.Join(*datadir, "neutrino.db")
	db, err := walletdb.Create("bdb", dbPath, true, time.Minute)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	cfg := neutrino.Config{
		DataDir:        *datadir,
		Database:       db,
		ChainParams:    params,
		ConnectPeers:   splitCSV(*connect),
		AddPeers:       splitCSV(*addpeer),
		DisableDNSSeed: *nodnsseed,
		DNSSeeds:       splitCSV(*dnsseed),
	}

	svc, err := neutrino.NewChainService(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "neutrino: %v\n", err)
		os.Exit(1)
	}
	if err := svc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start: %v\n", err)
		os.Exit(1)
	}
	defer svc.Stop()

	fmt.Println("syncing headers (Ctrl-C to stop)")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			tip, err := svc.BestBlock()
			if err != nil {
				fmt.Fprintf(os.Stderr, "tip: %v\n", err)
				continue
			}
			fmt.Printf("height=%d hash=%s current=%v peers=%d\n",
				tip.Height, tip.Hash, svc.IsCurrent(), len(svc.Peers()))
		}
	}
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
