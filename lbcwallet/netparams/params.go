// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package netparams

import (
	"github.com/lbryio/lbcd/chaincfg"
)

// Params is used to group parameters for various networks such as the main
// network and test networks.
type Params struct {
	*chaincfg.Params
	RPCClientPort string
	RPCServerPort string
}

// MainNetParams contains parameters specific running lbcwallet and
//
//	on the main network (wire.MainNet).
var MainNetParams = Params{
	Params:        &chaincfg.MainNetParams,
	RPCClientPort: "9245",
	RPCServerPort: "9244",
}

// TestNet3Params contains parameters specific running lbcwallet and
//
//	on the test network (version 3) (wire.TestNet3).
var TestNet3Params = Params{
	Params:        &chaincfg.TestNet3Params,
	RPCClientPort: "19245",
	RPCServerPort: "19244",
}

// RegNetParams contains parameters specific to the regression test network
// (wire.RegNet).
var RegTestParams = Params{
	Params:        &chaincfg.RegressionNetParams,
	RPCClientPort: "29245",
	RPCServerPort: "29244",
}
