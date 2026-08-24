// Command cfprobe dials LBC P2P peers and reports whether they advertise
// compact filters (SFNodeCF). Pass --connect or use DNS seeds from chain params.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lbryio/lbcd/chaincfg"
	"github.com/lbryio/lbcd/chaincfg/chainhash"
	"github.com/lbryio/lbcd/peer"
	"github.com/lbryio/lbcd/wire"
)

func main() {
	network := flag.String("network", "mainnet", "mainnet, testnet, or regtest")
	connect := flag.String("connect", "", "comma-separated host:port (skip DNS if set)")
	dnsseed := flag.String("dnsseed", "", "comma-separated DNS seeds (replaces defaults if set)")
	timeout := flag.Duration("timeout", 8*time.Second, "per-peer handshake timeout")
	flag.Parse()

	params := &chaincfg.MainNetParams
	switch strings.ToLower(*network) {
	case "mainnet", "main":
		params = &chaincfg.MainNetParams
	case "testnet", "testnet3":
		params = &chaincfg.TestNet3Params
	case "regtest", "simnet":
		params = &chaincfg.RegressionNetParams
	default:
		fmt.Fprintf(os.Stderr, "unknown network %q\n", *network)
		os.Exit(2)
	}

	targets := splitCSV(*connect)
	if len(targets) == 0 {
		seeds := params.DNSSeeds
		if extra := splitCSV(*dnsseed); len(extra) > 0 {
			seeds = nil
			for _, h := range extra {
				seeds = append(seeds, chaincfg.DNSSeed{Host: h})
			}
		}
		for _, s := range seeds {
			addrs, err := net.LookupHost(s.Host)
			if err != nil {
				fmt.Fprintf(os.Stderr, "seed %s: %v\n", s.Host, err)
				continue
			}
			for _, a := range addrs {
				targets = append(targets, net.JoinHostPort(a, params.DefaultPort))
			}
		}
	}
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "no peers: pass --connect host:port or check DNS seeds")
		os.Exit(1)
	}

	var (
		mu     sync.Mutex
		cf, n  int
	)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)
	for _, addr := range targets {
		addr := addr
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			ok, services, err := probe(params, addr, *timeout)
			mu.Lock()
			defer mu.Unlock()
			n++
			if err != nil {
				fmt.Printf("%s  error: %v\n", addr, err)
				return
			}
			flag := "-"
			if ok {
				cf++
				flag = "SFNodeCF"
			}
			fmt.Printf("%s  services=%s %s\n", addr, services, flag)
		}()
	}
	wg.Wait()
	fmt.Printf("\n%d/%d peers advertised compact filters (SFNodeCF)\n", cf, n)
	if cf == 0 {
		fmt.Fprintln(os.Stderr, "No compact-filter peers found. Point --connect at your own CF-capable daemon.")
		os.Exit(1)
	}
}

func probe(params *chaincfg.Params, addr string, timeout time.Duration) (bool, wire.ServiceFlag, error) {
	done := make(chan wire.ServiceFlag, 1)
	cfg := &peer.Config{
		UserAgentName:    "lbc-cfprobe",
		UserAgentVersion: "0.1.0",
		ChainParams:      params,
		Services:         0,
		TrickleInterval:  time.Second,
		NewestBlock: func() (*chainhash.Hash, int32, error) {
			return params.GenesisHash, 0, nil
		},
		Listeners: peer.MessageListeners{
			OnVersion: func(p *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
				select {
				case done <- msg.Services:
				default:
				}
				return nil
			},
		},
	}
	p, err := peer.NewOutboundPeer(cfg, addr)
	if err != nil {
		return false, 0, err
	}
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false, 0, err
	}
	p.AssociateConnection(conn)
	defer p.Disconnect()

	select {
	case services := <-done:
		return services&wire.SFNodeCF != 0, services, nil
	case <-time.After(timeout):
		return false, 0, fmt.Errorf("handshake timeout")
	}
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
