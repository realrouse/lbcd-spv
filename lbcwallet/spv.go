package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcwallet/walletdb"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/lbryio/lbcwallet/chain"
	"github.com/lbryio/lbcwallet/rpc/legacyrpc"
	"github.com/lbryio/lbcwallet/wallet"
	"github.com/lbc-spv/neutrino"
)

// spvConnectLoop starts Neutrino and attaches it to the loaded wallet. Unlike
// the RPC loop it does not reconnect to a full node; P2P reconnect is inside
// Neutrino.
func spvConnectLoop(legacyRPCServer *legacyrpc.Server, loader *wallet.Loader) {
	chainClient, err := startNeutrino()
	if err != nil {
		log.Errorf("Unable to start SPV chain service: %v", err)
		return
	}

	loader.RunAfterLoad(func(w *wallet.Wallet) {
		chainClient.SetStartTime(w.Manager.Birthday())
		w.SynchronizeRPC(chainClient)
		if legacyRPCServer != nil {
			legacyRPCServer.SetChainServer(chainClient)
		}
	})

	if err := chainClient.Start(); err != nil {
		log.Errorf("Unable to start Neutrino client: %v", err)
		return
	}

	addInterruptHandler(func() {
		chainClient.Stop()
		if err := chainClient.CS.Stop(); err != nil {
			log.Errorf("Error stopping Neutrino: %v", err)
		}
		chainClient.WaitForShutdown()
	})
}

func startNeutrino() (*chain.NeutrinoClient, error) {
	dataDir := networkDir(cfg.AppDataDir.Value, activeNet.Params)
	if err := checkCreateDir(dataDir); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dataDir, "neutrino.db")
	timeout := cfg.DBTimeout
	if timeout == 0 {
		timeout = time.Minute
	}
	db, err := walletdb.Create("bdb", dbPath, true, timeout)
	if err != nil {
		return nil, fmt.Errorf("opening neutrino.db: %w", err)
	}

	ncfg := neutrino.Config{
		DataDir:        dataDir,
		Database:       db,
		ChainParams:    *activeNet.Params,
		ConnectPeers:   cfg.ConnectPeers,
		AddPeers:       cfg.AddPeers,
		DisableDNSSeed: cfg.NoDNSSeed,
		DNSSeeds:       cfg.DNSSeeds,
	}
	svc, err := neutrino.NewChainService(ncfg)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	log.Infof("SPV Neutrino using datadir %s (peers=%v nodnsseed=%v)",
		dataDir, append(append([]string{}, cfg.ConnectPeers...), cfg.AddPeers...), cfg.NoDNSSeed)
	return chain.NewNeutrinoClient(activeNet.Params, svc), nil
}
