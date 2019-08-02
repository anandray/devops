// Copyright (c) 2018 Wolk Inc.  All rights reserved.

// The SWARMDB library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The SWARMDB library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package wolk

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"

	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
)

type GenesisConfig struct {
	NetworkID     uint64                     `json:"networkID,omitempty"`
	Seed          []byte                     `json:"seed,omitempty"`
	DeltaPos      uint64                     `json:"deltaPos,omitempty"`
	DeltaNeg      uint64                     `json:"deltaNeg,omitempty"`
	StorageBeta   uint64                     `json:"storageBeta,omitempty"`
	BandwidthBeta uint64                     `json:"bandwidthBeta,omitempty"`
	PhaseVotes    uint64                     `json:"phaseVotes,omitempty"`
	Q             uint8                      `json:"q,omitempty"`
	Gamma         uint64                     `json:"gamma,omitempty"`
	Accounts      map[common.Address]Account `json:"accounts,omitempty"`
	Registry      []SerializedNode           `json:"nodes,omitempty"`
}

const (
	defaultDomain = "wolk.com"
)

func (self *GenesisConfig) Dump() {
	for i, r := range self.Registry {
		log.Info("[genesis:Dump]", "i", i, "node", r.String())
	}
}

//return a pre-genesis block that contains block parameters used in stateDB
func (c *GenesisConfig) CreatePreGenesis() *Block {
	b := new(Block)
	b.NetworkID = c.NetworkID
	b.Seed = c.Seed
	b.StorageBeta = c.StorageBeta
	b.BandwidthBeta = c.BandwidthBeta
	return b
}

func CreateGenesisFile(networkID int, filename string, accounts map[common.Address]Account, registry []SerializedNode) (err error) {
	var c GenesisConfig
	c.NetworkID = uint64(networkID)
	c.StorageBeta = 30
	c.BandwidthBeta = 40
	c.Accounts = accounts
	c.Registry = registry
	str := fmt.Sprintf("%d", c.NetworkID)
	c.Seed = wolkcommon.Computehash([]byte(str))
	return c.Save(filename)
}

// save file
func (c *GenesisConfig) Save(filename string) (err error) {
	cout, err1 := json.MarshalIndent(c, "", "\t")
	if err1 != nil {
		return err1
	}
	err = ioutil.WriteFile(filename, cout, 0644)
	if err != nil {
		return err
	}
	return nil
}

func LoadGenesisFile(filename string) (c *GenesisConfig, err error) {
	// read file
	c = new(GenesisConfig)
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(dat, c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (self *GenesisConfig) getip(n string) (ip string, err error) {
	ips, err := net.LookupIP(n)
	if err != nil {
		return ip, err
	}
	for _, ipr := range ips {
		return fmt.Sprintf("%s", ipr), nil
	}
	return ip, fmt.Errorf("Not Found")
}

func (self *GenesisConfig) GetTrustedNodes(offset int, trustedNode string) (nodes []*enode.Node, err error) {
	nodes = make([]*enode.Node, 0)
	for i, r := range self.Registry {
		if strings.Compare(trustedNode, r.ConsensusIP) == 0 {
			key := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))
			epkey, _ := ethcrypto.HexToECDSA(key)
			pubkey := &epkey.PublicKey
			ip, err := self.getip(r.ConsensusIP)
			if err == nil {
				log.Debug("[genesis:GetTrustedNodes] getip", "i", i, "ip", ip)
				node := enode.NewV4(pubkey, net.ParseIP(ip), int(30303+offset*1000), int(30303+offset*1000))
				nodes = append(nodes, node)
			}
		}
	}
	return nodes, nil
}

func LoadGenesisBayes(networkID uint64, instanceGroup string) (c *GenesisConfig, idx int, identity string, nodeType string, ecdsaKey *ecdsa.PrivateKey, edwardsKey *crypto.PrivateKey, err error) {

	c = new(GenesisConfig)
	c.NetworkID = uint64(networkID)
	c.Accounts = make(map[common.Address]Account)
	c.Registry = make([]SerializedNode, 0)
	str := fmt.Sprintf("%d", c.NetworkID)
	c.Seed = wolkcommon.Computehash([]byte(str))

	// not used yet...
	c.DeltaPos = 1
	c.DeltaNeg = 2
	c.StorageBeta = 30
	c.BandwidthBeta = 40
	c.PhaseVotes = 0
	c.Q = 3
	c.Gamma = 100

	// set up accounts
	addr := common.HexToAddress("0xdA2741Fee2928Eaa03Ba68007001a2b046Dd4045")
	acc := Account{Balance: uint64(1000)}
	c.Accounts[addr] = acc

	// set up idenity variables
	nodeType = "consensus"
	identity, _ = os.Hostname()

	dbURI := fmt.Sprintf("root:1wasb0rn2!@tcp(db03)/wolk")
	db, err := sql.Open("mysql", dbURI)
	if err != nil {
		log.Error("err", err)
		return c, idx, identity, nodeType, ecdsaKey, edwardsKey, err
	}
	rows, err := db.Query(fmt.Sprintf("SELECT hostname, dns, publicip from algorand where instanceGroup = \"%s\" order by hostname", instanceGroup))
	if err != nil {
		log.Error("[genesis:LoadGenesisBayes] Query", "err", err)
		return c, idx, identity, nodeType, ecdsaKey, edwardsKey, err
	}
	defer rows.Close()
	var hostname string
	var dns string
	var publicip string

	for i := 0; rows.Next(); i++ {
		err = rows.Scan(&hostname, &dns, &publicip)
		if err != nil {
			log.Error("[genesis:LoadGenesisBayes] Scan", "err", err)
			return c, idx, identity, nodeType, ecdsaKey, edwardsKey, err
		}
		privateKeyStr := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))
		ecdsaPrivateKey, _ := ethcrypto.HexToECDSA(privateKeyStr)
		var n = new(SerializedNode)
		n.Address = addr
		n.PubKey = fmt.Sprintf("%x", ethcrypto.CompressPubkey(&ecdsaPrivateKey.PublicKey))
		n.ValueInt = 1000
		n.ValueExt = 100
		n.StorageIP = fmt.Sprintf("%s.%s", dns, defaultDomain)
		n.ConsensusIP = fmt.Sprintf("%s.%s", dns, defaultDomain)
		n.Region = RegionNA
		n.HTTPPort = 443
		c.Registry = append(c.Registry, *n)
		if strings.Compare(hostname, identity) == 0 {
			idx = i
			fmt.Printf(" ***** ")
			ecdsaKey = ecdsaPrivateKey
			edwardsKey, _ = crypto.HexToPrivateKey(privateKeyStr)
		}
	}
	return c, idx, identity, nodeType, ecdsaKey, edwardsKey, err
}

func (self *GenesisConfig) GetStaticNodes(offset int) (nodes []*enode.Node, err error) {
	nodes = make([]*enode.Node, 0)
	for i, r := range self.Registry {

		key := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))
		epkey, _ := ethcrypto.HexToECDSA(key)
		pubkey := &epkey.PublicKey

		ip, err := self.getip(r.ConsensusIP)
		if err == nil {
			log.Info("GetStaticNodes", "i", i, "inp", r.ConsensusIP, "ip", ip)
			node := enode.NewV4(pubkey, net.ParseIP(ip), int(30303+offset*1000), int(30303+offset*1000))
			nodes = append(nodes, node)
		} else {
			log.Error("getip", "err", err)
		}
	}
	return nodes, nil
}

// get static node to peer with the storage node (they must have the same ConsensusIdx)
func (self *GenesisConfig) GetStaticNode(offset int, consensusIdx int) (nodes []*enode.Node, err error) {
	nodes = make([]*enode.Node, 0)

	trueConsensusIdx := consensusIdx % 8 //TODO: make it work better with registry
	if len(self.Registry) >= trueConsensusIdx {
		return nodes, fmt.Errorf("[genesis:GetStorageNode] invalid consensusIdx requested")
	}
	r := self.Registry[trueConsensusIdx]

	key := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", trueConsensusIdx))))
	epkey, _ := ethcrypto.HexToECDSA(key)
	pubkey := &epkey.PublicKey
	ip, err := self.getip(r.ConsensusIP)
	if err == nil {
		log.Debug("[genesis:GetStaticNodes] getip", "ip", ip)
		node := enode.NewV4(pubkey, net.ParseIP(ip), int(30303+offset*1000), int(30303+offset*1000))
		nodes = append(nodes, node)
		log.Trace("[genesis:GetStaticNode] peer storage and consensus node with the same consensusIdx", "consensusIdx", consensusIdx)
	} else {
		log.Trace("[genesis:GetStaticNode] FAIL to peer storage and consensus node with the same consensusIdx", "consensusIdx", consensusIdx)
	}
	return nodes, nil
}

func (self *GenesisConfig) GetStorageNode(offset int, consensusIdx int) (nodes []*enode.Node, err error) {
	nodes = make([]*enode.Node, 0)
	log.Debug(fmt.Sprintf("Registry is of length: %d contents: %+v", len(self.Registry), self.Registry))

	trueConsensusIdx := consensusIdx % 8 //TODO: make it work better with registry
	if len(self.Registry) >= trueConsensusIdx {
		return nodes, fmt.Errorf("[genesis:GetStorageNode] invalid consensusIdx requested")
	}
	r := self.Registry[trueConsensusIdx]
	key := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", trueConsensusIdx))))
	epkey, _ := ethcrypto.HexToECDSA(key)
	pubkey := &epkey.PublicKey
	/*
	   pubkey, err := crypto.UnmarshalPubkey(common.FromHex(r.PubKey))
	   if err != nil {
	           return nodes, err
	   }
	*/
	ip, err := self.getip(r.StorageIP)
	if err == nil {
		log.Debug("[genesis:GetStaticNodes] getip", "ip", ip)
		node := enode.NewV4(pubkey, net.ParseIP(ip), int(30303+offset*1000), int(30303+offset*1000))
		nodes = append(nodes, node)
		log.Trace("[genesis:GetStorageNode] peer storage and consensus node with the same consensusIdx", "consensusIdx", consensusIdx)
	} else {
		log.Trace("[genesis:GetStorageNode] FAIL to peer storage and consensus node with the same consensusIdx", "consensusIdx", consensusIdx)
	}
	return nodes, nil
}
