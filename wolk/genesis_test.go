package wolk

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/wolkdb/cloudstore/log"
)

func getAccounts() map[common.Address]Account {
	addr := common.HexToAddress("0xdA2741Fee2928Eaa03Ba68007001a2b046Dd4045")
	acc := Account{Balance: uint64(1000)}
	accounts := make(map[common.Address]Account)
	accounts[addr] = acc
	return accounts
}

func getRegisteredNode() []SerializedNode {
	bal := uint64(10000)
	res := []SerializedNode{
		{Address: common.HexToAddress("0xdA2741Fee2928Eaa03Ba68007001a2b046Dd4045"), PubKey: "03d67104e65b61fca1fc73a74175667c11f16c037fa8e414499189df9d68706aca", StorageIP: "35.190.55.204", ConsensusIP: "35.199.185.185", ValueInt: bal, Region: RegionNA},   // wolk-0-gc-na-datastore-0
		{Address: common.HexToAddress("0x2d77B4b82F575bC8020445d92bBe09F2d0b0E8a6"), PubKey: "036d11b37ef4fc909ae1ec6c63584f99843f04ea9466730f8e58950c8c7262e693", StorageIP: "107.178.255.219", ConsensusIP: "35.245.222.192", ValueInt: bal, Region: RegionNA}, // wolk-15-gc-na-datastore-1
		{Address: common.HexToAddress("0xe1fbe354b0fdc8f5e67a616cffe386f9a700bfae"), PubKey: "0294dd98401cf6ca0418580bb77ad0606a35ae8f442b22ab0f81011d8d2c78e70f", StorageIP: "35.227.250.248", ConsensusIP: "35.232.19.5", ValueInt: bal, Region: RegionNA},     // wolk-2-gc-na-bigtable-2
		{Address: common.HexToAddress("0xf92241462abaafde14acedf72c70e94d8b6788d0"), PubKey: "02d9d5b902798f5c42901f3fd993e58a9fb2ded12c3996e2a3849d4fdfcd0c3516", StorageIP: "35.227.234.240", ConsensusIP: "35.201.254.135", ValueInt: bal, Region: RegionAP},  // wolk-3-gc-as-datastore-3
		{Address: common.HexToAddress("0xc38c8ae2cf701e38ea2a77d881e997eedff099cb"), PubKey: "03ae6aee36993a5b48cf3a36a6cf10cf5e767e56e253d232f8462a4b74ee3374d0", StorageIP: "35.244.184.89", ConsensusIP: "35.246.154.85", ValueInt: bal, Region: RegionEU},    // wolk-4-gc-eu-datastore-4
		{Address: common.HexToAddress("0x985098b282d73a175a30f8160bb225b0f221d987"), PubKey: "03142e89f6535bc36c05215eec0938234c06e91cb9b0e7977706b72775463b252b", StorageIP: "130.211.28.29", ConsensusIP: "35.185.154.5", ValueInt: bal, Region: RegionAS},     // wolk-5-gc-in-datastore-5
		{Address: common.HexToAddress("0xda2741fee2928eaa03ba68007001a2b046dd4045"), PubKey: "0314f2a4e39998a282636f1ee1e8e4ed29bd2b1285f9dad01837ac1453ea43d648", StorageIP: "35.190.55.204", ConsensusIP: "35.247.55.13", ValueInt: bal, Region: RegionNA},     // wolk-0-gc-na-datastore-6
		{Address: common.HexToAddress("0xda2741fee2928eaa03ba68007001a2b046dd4045"), PubKey: "02afcb05e7b74b3124443f49044ea8c2d94f06824dee0eda489806b98c62673d01", StorageIP: "35.227.250.248", ConsensusIP: "35.232.47.23", ValueInt: bal, Region: RegionNA},    //wolk-2-gc-na-bigtable-7
	}
	return res
}

func TestProduction(t *testing.T) {
	filename := "./cloud/credentials/genesisTEST.json"
	genesis, err := LoadGenesisFile(filename)
	if err != nil {
		t.Fatalf("LoadGenesisFile %v\n", err)
	}
	genesis.Dump()
	nodes, err := genesis.GetStaticNodes(0)
	if err != nil {
		t.Fatalf("GetStaticNodes %v\n", err)
	}
	log.Info("GetStaticNodes", "len(nodes)", len(nodes))
}

// run this test to build a new genesis file
func TestGenesis(t *testing.T) {
	registry := getRegisteredNode()
	accounts := getAccounts()
	filename := "./cloud/credentials/genesisTEST.json"
	networkID := int(1234)
	err := CreateGenesisFile(networkID, filename, accounts, registry)
	if err != nil {
		t.Fatalf("CreateGenesisFile %v\n", err)
	}
	genesis, err := LoadGenesisFile(filename)
	if err != nil {
		t.Fatalf("LoadGenesisFile %v\n", err)
	}
	for i, rn := range registry {
		a := genesis.Registry[i]
		if bytes.Compare(rn.Address.Bytes(), a.Address.Bytes()) != 0 {
			t.Fatalf("Address Mismatch %s != %s", rn.Address, a.Address)
		}
		if strings.Compare(rn.PubKey, a.PubKey) != 0 {
			t.Fatalf("Address Mismatch %s != %s", rn.Address, a.Address)
		}
		if rn.ValueInt != a.ValueInt {
			t.Fatalf("ValueInt Mismatch %d != %d", rn.ValueInt, a.ValueInt)
		}
		if rn.ValueExt != a.ValueExt {
			t.Fatalf("ValueExt Mismatch %d != %d", rn.ValueExt, a.ValueExt)
		}
		if rn.Region != a.Region {
			t.Fatalf("Region Mismatch %d != %d", rn.Region, a.Region)
		}
		if strings.Compare(rn.StorageIP, a.StorageIP) != 0 {
			t.Fatalf("IP Mismatch %s != %s", rn.StorageIP, a.StorageIP)
		}
		if strings.Compare(rn.ConsensusIP, a.ConsensusIP) != 0 {
			t.Fatalf("IP Mismatch %s != %s", rn.ConsensusIP, a.ConsensusIP)
		}
		fmt.Printf("PASS %v\n", a)
	}
}
