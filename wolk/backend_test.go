// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	rand "math/rand"
	"net"
	"net/http"
	"path"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
	"github.com/wolkdb/cloudstore/log"
	"github.com/wolkdb/cloudstore/wolk/cloud"
)

const (
	defaultNetworkID int = 1234
	genesisFileName      = "genesis.json"
	p2pPort              = 30303
)

func matched_chunk(c1 []byte, c2 []byte, l int) bool {
	return bytes.Compare(c1[0:l], c2[0:l]) == 0
}

func generateRandomData(l int) (r io.Reader, slice []byte) {
	slice = make([]byte, l)
	rand.Seed(time.Now().Unix())
	if _, err := rand.Read(slice); err != nil {
		panic("rand error")
	}
	r = io.LimitReader(bytes.NewReader(slice), int64(l))
	return
}

func RegisterWolkService(stack *node.Node, cfg *cloud.Config, t *testing.T) <-chan *WolkStore {
	ch := make(chan *WolkStore, 1)
	var err error
	err = stack.Register(func(ctx *node.ServiceContext) (w node.Service, err error) {
		wcs, err := NewWolk(nil, cfg)
		if err != nil {
			log.Error("ERR2", "err", err)
			return w, err
		}
		ch <- wcs
		return wcs, err
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Trace(fmt.Sprintf("[backend_test:RegisterWolkService] RegisterWolkService SUCCESS %v", ch))
	return ch
}

func ReleaseNodes(t *testing.T, n [MAX_VALIDATORS]*node.Node) {
	for _, node := range n {
		node.Stop()
	}
}

func newWolkPeer(t *testing.T, consensusAlgo string) (wolk []*WolkStore, privateKeysECDSA []*ecdsa.PrivateKey, accounts map[common.Address]Account, registry []SerializedNode, n [MAX_VALIDATORS]*node.Node) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	wolk = make([]*WolkStore, MAX_VALIDATORS)
	nAccounts := MAX_VALIDATORS // TODO: BUG FIX -- this is a requirement right now because counting is done by address signing proposals

	log.Trace("Selected Consensus", "algo", consensusAlgo)

	privateKeys := make([]*crypto.PrivateKey, nAccounts)
	pubKeys := make([]string, nAccounts)
	privateKeysECDSA = make([]*ecdsa.PrivateKey, nAccounts)
	pubKeysECDSA := make([]string, nAccounts)
	accounts = make(map[common.Address]Account)
	addr := make([]common.Address, nAccounts)
	for i := 0; i < nAccounts; i++ {
		k_str := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", i))))
		var err error
		privateKeys[i], err = crypto.HexToPrivateKey(k_str)
		if err != nil {
			t.Fatalf("[backend_test:newWolkPeer] HexToEdwardsPrivateKey %v", err)
		}

		// TODO:
		// this wont work for p2p peering, unless you make a change in the registry[i].PubKey line ( which is Edwards based ) --
		// privateKeysECDSA[i], err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

		// this is necessary for p2p to work since the public keys in the genesis file match up with these determinstic keys --
		privateKeysECDSA[i], err = ethcrypto.HexToECDSA(k_str)
		if err != nil {
			t.Fatalf("[backend_test:newWolkPeer] HexToECDSA %v", err)
		}
		pubKeys[i] = fmt.Sprintf("%x", privateKeys[i].PublicKey())
		pubKeysECDSA[i] = fmt.Sprintf("%x", ethcrypto.CompressPubkey(&privateKeysECDSA[i].PublicKey))
		//addr[i] = crypto.PubkeyToAddress(privateKeys[i].PublicKey())
		addr[i] = crypto.GetECDSAAddress(privateKeysECDSA[i]) //MUST FIX
		accounts[addr[i]] = Account{Balance: uint64(100000 + i)}
	}
	registry = make([]SerializedNode, MAX_VALIDATORS)
	for i := 0; i < MAX_VALIDATORS; i++ {
		p, err := ethcrypto.DecompressPubkey(common.FromHex(pubKeysECDSA[i]))
		if err != nil {
			t.Fatalf("Problem! %v", err)
		}
		registry[i] = SerializedNode{
			Address:     addr[i],
			PubKey:      pubKeys[i],
			ValueInt:    uint64(10000 + i),
			ValueExt:    uint64(0),
			StorageIP:   "127.0.0.1",
			ConsensusIP: "127.0.0.1",
			Region:      1,
			HTTPPort:    uint16(httpPort + i),
		}
		tmpnode := enode.NewV4(p, net.ParseIP(registry[i].ConsensusIP), 30301, 30303)
		log.Trace("[backend_test:newWolkPeer] KeyGeneration", "i", i, "addr", addr[i], "enode", tmpnode)
	}
	err := CreateGenesisFile(defaultNetworkID, genesisFileName, accounts, registry)
	if err != nil {
		t.Fatalf("CreateGenesisFile err: %+v\n", err)
	}
	genesisConfig, err := LoadGenesisFile(genesisFileName)
	if err != nil {
		t.Fatalf("LoadGenesisFile err: %+v\n", err)
	}

	str := fmt.Sprintf("%d", int32(time.Now().Unix()))
	if useResume {
		str = fmt.Sprintf("%v", consensusAlgo)
	}

	/*
		dir, err := ioutil.ReadDir("/tmp")
		for _, d := range dir {
				os.RemoveAll(path.Join([]string{"tmp", d.Name()}...))
		}
	*/

	// create nodes
	var nodelist [MAX_VALIDATORS]*node.Node
	for i := 0; i < MAX_VALIDATORS; {
		registeredNode := registry[i]
		nodecfg := &node.DefaultConfig
		var err error
		nodecfg.P2P.StaticNodes, err = genesisConfig.GetStaticNodes(0)
		if err != nil {
			t.Fatalf("GetStaticNodes err %v", err)
		}
		nodecfg.P2P.ListenAddr = fmt.Sprintf(":%d", 10000+i*32) // TODO
		nodecfg.DataDir = fmt.Sprintf("/tmp/node%s/datadir%d", str, i)
		nodecfg.HTTPHost = registeredNode.ConsensusIP
		nodecfg.HTTPPort = 9900 + i
		nodecfg.HTTPModules = append(nodecfg.HTTPModules, "admin")
		nodecfg.UserIdent = fmt.Sprintf("%v%d", "consensus", i)
		nodecfg.Name = "wolk"
		stack, err := node.New(nodecfg)
		if err != nil {
			log.Error("[backend_test:newWolkPeer] node.New FAILURE %v", err)
			//	t.Fatal(err)
		} else {
			log.Trace("[backend_test:newWolkPeer] node.New SUCC", "i", i)
		}
		datadir := fmt.Sprintf("/tmp/consensus%s/datadir%d", str, i)
		cfg := cloud.DefaultConfig
		cfg.GenesisFile = genesisFileName
		cfg.DataDir = datadir
		cfg.HTTPPort = int(registeredNode.HTTPPort)
		cfg.ConsensusIdx = i
		cfg.NodeType = "consensus"
		cfg.ConsensusAlgorithm = consensusAlgo
		cfg.Preemptive = usePreemptive
		cfg.Provider = "leveldb"
		cfg.Address = registeredNode.Address
		cfg.OperatorKey = privateKeys[i%nAccounts]
		cfg.OperatorECDSAKey = privateKeysECDSA[i%nAccounts]

		RegisterWolkService(stack, &cfg, t)
		nodelist[i] = stack
		done := false
		for done == false {
			err = nodelist[i].Start()
			if err != nil {
				log.Error("[backend_test:newWolkPeer] node START failure", "err", err)
			} else {
				svr := nodelist[i].Server()
				if svr != nil {
					log.Info("[backend_test:newWolkPeer] Created node", "i", i)
					done = true
					i++
				} else {
					log.Error("[backend_test:newWolkPeer] Could not create node", "i", i, "p2pPort", nodecfg.P2P.ListenAddr, "err", err, "datadir", datadir)
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	// set up peering between nodes (each peered with 3 others in a ring)
	// set up peering between nodes (each peered with 3 others in a ring)

	for i := 0; i < MAX_VALIDATORS; i++ {
		for j := i; j < MAX_VALIDATORS; j++ {
			svr := nodelist[i].Server() // definitely there
			s := nodelist[j].Server()
			q := s.Self()
			svr.AddPeer(q)
		}
	}

	// for all the nodes, get their blockchains
	success := 0
	for i := 0; i < MAX_VALIDATORS; i++ {
		var wcs *WolkStore
		err = nodelist[i].Service(&wcs)
		if err != nil {
			log.Error("[backend_test:newWolkPeer] SERVICE", "err", err)
		} else {
			success++
		}
		wolk[i] = wcs
		// wcs.Start()
	}
	if success < MAX_VALIDATORS {
		t.Fatalf("[backend_test:newWolkPeer] Insufficent nodes")
	}

	done := false
	for node := 0; !done; {
		if !wolk[node].isChainReady() {
			time.Sleep(400 * time.Millisecond)
		} else {
			node++
			if node == MAX_VALIDATORS {
				done = true
			}
		}
	}

	log.Info("[backend_test:newWolkPeer] Wolk Peer Setup complete")
	return wolk, privateKeysECDSA, accounts, registry, nodelist
}

func doTransaction(t *testing.T, ws *WolkStore, tx *Transaction) *Transaction {
	_, err := ws.SendRawTransaction(tx)
	if err != nil {
		t.Fatal(err)
	}
	// allow time for gossiping, producing a block
	for i := 0; i < 30; i++ {
		time.Sleep(1000 * time.Millisecond)
		tx2, bn, _, ok, err := ws.GetTransaction(context.TODO(), tx.Hash())
		if err != nil || !ok {
			log.Info(fmt.Sprintf("Not included yet: tx(%x)  [Node%d]", tx.Hash(), ws.consensusIdx))
		} else if ok && bn > 0 {
			log.Warn(fmt.Sprintf("TRANSACTION Included in %d: tx(%x)  [Node%d] %s", bn, tx.Hash(), ws.consensusIdx, tx.String()))
			return tx2
		}
	}
	t.Fatalf("[backend_test:doTransaction] tx(%x) did not get included! tried polling for it 30 times.", tx.Hash())
	return tx
}

func waitForTxDone(t *testing.T, txhash common.Hash, wolkStore *WolkStore) {
	log.Trace("[backend_nosql_test:waitForTxsDone]", "tx", txhash)
	time.Sleep(500 * time.Millisecond)
	maxtries := 30
	gottx := false
	try := 0
	for gottx == false && try < maxtries {
		_, bn, _, ok, err := wolkStore.GetTransaction(context.TODO(), txhash)
		if err != nil || !ok {
			if err != nil {
				tprint("[waitForTxsDone] tx(%x), err(%s)", txhash, err)
			}
			time.Sleep(1000 * time.Millisecond)
			//tprint("SendRawTransaction (%x) not done yet...", txhash)
		} else if bn > 0 {
			gottx = true
			tprint("[waitForTxDone] got it")
		} else {
			time.Sleep(1000 * time.Millisecond)
		}
		try++
	}
	if gottx == false {
		t.Fatalf("tried 30 times and tx(%x) was not found", txhash)
	}
}

func waitForTxsDone(t *testing.T, txhashes []common.Hash, wolkStore *WolkStore) {
	log.Trace("[waitForTxsDone] waiting")
	maxtries := 30
	var gottx bool
	var try int

	for _, txhash := range txhashes {
		time.Sleep(500 * time.Millisecond)
		gottx = false
		try = 0
		for gottx == false && try < maxtries {
			_, bn, _, ok, err := wolkStore.GetTransaction(context.TODO(), txhash)
			if err != nil || !ok {
				if err != nil {
					log.Error("[waitForTxsDone]", "txhash", txhash, "err", err)
				}
				//tprint("SendRawTransaction (%x) not done yet...", txhash)
				time.Sleep(1000 * time.Millisecond)
			} else if bn > 0 {
				gottx = true
				log.Trace("[waitForTxsDone] FOUND", "txhash", txhash)
			} else {
				time.Sleep(1000 * time.Millisecond)
			}
			try++
		}
		if gottx == false {
			t.Fatalf("[waitForTxsDone] TX NOT FOUND %x try: %d / %d", txhash, try, maxtries)
		}
	}
	log.Trace("[waitForTxsDone] DONE")
	return
}

func TestMultiTransfers1(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, privateKeys, _, _, _ := newWolkPeer(t, "singlenode")
	// defer ReleaseNodes(t, nodelist)
	time.Sleep(25 * time.Second)

	// submit the SetName transaction
	node := 0
	if wolk[node].NumPeers() < 2 {
		t.Fatalf("Insufficient peering relationships setup in blockchain")
	}

	txaddr := make(map[common.Hash]common.Address)
	txowner := make(map[common.Hash]string)

	txlist := []*Transaction{}

	for ownernumber := 0; ownernumber < MAX_VALIDATORS; ownernumber++ {
		owner := fmt.Sprintf("owner%d", ownernumber)
		pk := privateKeys[ownernumber%MAX_VALIDATORS]
		addr := crypto.GetECDSAAddress(pk)
		log.Info("SetName---INIT", "owner", owner)
		tx, err := NewTransaction(pk, http.MethodPut, owner, NewTxAccount(owner, []byte("FakeRSAPublicKeyForAccounts")))
		if err != nil {
			t.Fatal(err)
		}

		txlist = append(txlist, tx)
		txhash := tx.Hash()
		txaddr[txhash] = addr
		txowner[txhash] = owner
	}

	txsresult := wolk[node].SendRawTransactions2(txlist)
	for txhash, err := range txsresult {
		if err != nil {
			log.Trace(fmt.Sprintf("Transaction error: tx(%x)  [Error%s]", txhash, err.Error()))
		}
	}

	for txhash, addr := range txaddr {
		owner := txowner[txhash]
		// initialize a map of which nodes have the tx
		hastx := make(map[int]bool)
		for i := 0; i < MAX_VALIDATORS; i++ {
			hastx[i] = false
		}
		start := time.Now()
		for done := false; !done; {
			count := 0
			for i := 0; i < MAX_VALIDATORS; i++ {
				if hastx[i] == false {
					_, bn, _, ok, err := wolk[i].GetTransaction(context.TODO(), txhash)
					if err != nil || !ok {
						log.Trace(fmt.Sprintf("Not included yet: tx(%x)  [Node%d] %s", txhash, i, time.Since(start)))
					} else if ok {
						log.Info(fmt.Sprintf("Included: tx(%x) in BLOCK %d [Node%d] %s", txhash, bn, i, time.Since(start)))
						hastx[i] = true
					}
				} else {
					count++
				}
			}
			log.Info(fmt.Sprintf("%d/%d nodes have SetName transaction for %s [%s]\n", count, MAX_VALIDATORS, owner, time.Since(start)))
			if count == MAX_VALIDATORS {
				done = true
			}
		}

		// do a getname call and see if it matches addr
		count := 0
		for i := 0; i < MAX_VALIDATORS; i++ {
			var options RequestOptions
			receivedaddr, ok, _, err := wolk[i].GetName(owner, &options)
			if err != nil {
				t.Fatalf("GetName ERR %v", err)
			} else if ok {
				log.Info("GetName Success", "node", i, "name", owner)
			} else {
				log.Info("GetName NOT FOUND", "node", i, "name", owner)
			}
			if bytes.Compare(addr.Bytes(), receivedaddr.Bytes()) != 0 {
				t.Fatalf("GetName INCORRECT addr %v != receivedaddr %v", addr, receivedaddr)
			} else {
				count++
			}
		}
		if count != MAX_VALIDATORS {
			t.Fatalf("INCONSISTENCY ACROSS nodes!")
		}
	}

	// TEST 1: addr1 is transferring 15 WOLK to addr0
	// CHECK:
	//  1. balance of address 0 should be 15 more
	//  2. balance of address 1 should 15 less!

	txlist = nil

	amount := uint64(15)
	txs := make([]*Transaction, MAX_VALIDATORS)
	balance_start := make([]uint64, MAX_VALIDATORS)
	addr0 := crypto.GetECDSAAddress(privateKeys[0])
	balance_start[0], _, _ = wolk[0].GetBalance(addr0, 0)
	for r := 1; r < MAX_VALIDATORS; r++ {

		addr := crypto.GetECDSAAddress(privateKeys[r])
		balance_start[r], _, _ = wolk[r].GetBalance(addr, 0)
		tx, err := NewTransaction(privateKeys[0], http.MethodPost, path.Join(ProtocolName, "transfer"), NewTxTransfer(amount, fmt.Sprintf("owner%d", r)))
		if err != nil {
			t.Fatalf("NewTransaction %v", err)
		}

		txlist = append(txlist, tx)
		txhash := tx.Hash()

		//// submit the Transfer transaction
		//txhash, err := wolk[(r+4)%MAX_VALIDATORS].SendRawTransaction(tx)
		//if err != nil {
		//	t.Fatal(err)
		//}
		txs[r] = tx
		log.Info("**** SENT TRANSFER TX ****", "txhash", txhash, "tx", tx)
	}

	txsresult = wolk[node].SendRawTransactions2(txlist)
	for txhash, err := range txsresult {
		if err != nil {
			log.Trace(fmt.Sprintf("Transaction error: tx(%x)  [Error%s]", txhash, err.Error()))
		}
	}

	for r := 1; r < MAX_VALIDATORS; r++ {
		tx := txs[r]
		hastx := make(map[int]bool)
		for i := 0; i < MAX_VALIDATORS; i++ {
			hastx[i] = false
		}
		start := time.Now()
		for done := false; !done; {
			missing := 0
			for i := 0; i < MAX_VALIDATORS; i++ {
				if hastx[i] == false {
					_, bn, _, ok, err := wolk[i].GetTransaction(context.TODO(), tx.Hash())
					if err != nil || !ok {
						log.Trace(fmt.Sprintf("Not included yet: tx(%x)  [Node%d] %s", tx.Hash(), i, time.Since(start)))
						missing++
					} else if ok {
						log.Trace(fmt.Sprintf("Included: tx(%x) in BLOCK %d [Node%d] %s", tx.Hash(), bn, i, time.Since(start)))
						hastx[i] = true
					}
				}
			}
			log.Info(fmt.Sprintf("%d/%d nodes are *missing* transfer transaction %d [%s]\n", missing, MAX_VALIDATORS, r, time.Since(start)))
			if missing == 0 {
				done = true
			} else {
				time.Sleep(1000 * time.Millisecond)
			}
		}
	}

	for i := 0; i < MAX_VALIDATORS; i++ {
		// do GetBalance calls and see if it matches
		balance0_end, ok, err := wolk[i].GetBalance(addr0, 0)
		if err != nil {
			t.Fatalf("GetBalance0 ERR %v", err)
		} else if ok {
			log.Info("GetBalance0 Success", "node", i, "addr", addr0, "balance", balance0_end)
		} else {
			log.Info("GetBalance NOT FOUND", "node", i, "addr", addr0)
		}
		expected_diff := int64(-amount) * (MAX_VALIDATORS - 1)
		diff := int64(balance0_end) - int64(balance_start[0])
		if diff != expected_diff {
			t.Fatalf("TX TEST Balance0 FAILURE %d != %d", diff, expected_diff)
		} else {
			log.Info("TX TEST Balance0 SUCCESS")
		}

		count := 0
		for r := 1; r < MAX_VALIDATORS; r++ {
			addr := crypto.GetECDSAAddress(privateKeys[r])
			balance_end, ok1, err := wolk[r].GetBalance(addr, 0)
			if err != nil {
				t.Fatalf("GetBalance1 ERR %v", err)
			} else if !ok1 {
				log.Info("GetBalance1 NOT FOUND", "node", i, "addr", addr)
			}
			diff := int64(balance_end) - int64(balance_start[r])
			if diff != int64(amount) {
				t.Fatalf("TX TEST Node%d Balance%d FAILURE %d != %d", i, r, diff, amount)
			} else {
				log.Info("TX TEST GetBalance SUCCESS", "Node", i, "Balance", balance_end, "addr", addr)
				count++
			}
		}
		if count != MAX_VALIDATORS-1 {
			t.Fatalf("INCONSISTENCY ACROSS nodes!")
		}
		log.Info("TX TEST CONSISTENT", "Node", i)
	}
}

func TestMultiTransfers2(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, privateKeys, _, _, _ := newWolkPeer(t, "singlenode")
	// defer ReleaseNodes(t, nodelist)
	time.Sleep(25 * time.Second)

	// submit the SetName transaction
	node := 0
	if wolk[node].NumPeers() < 2 {
		t.Fatalf("Insufficient peering relationships setup in blockchain")
	}

	txaddr := make(map[common.Hash]common.Address)
	txowner := make(map[common.Hash]string)

	txlist := []*Transaction{}

	for ownernumber := 0; ownernumber < MAX_VALIDATORS; ownernumber++ {
		owner := fmt.Sprintf("owner%d", ownernumber)
		pk := privateKeys[ownernumber%MAX_VALIDATORS]
		addr := crypto.GetECDSAAddress(pk)
		log.Info("SetName---INIT", "owner", owner)
		tx, err := NewTransaction(pk, http.MethodPut, owner, NewTxAccount(owner, []byte("FakeRSAPublicKeyForAccounts")))
		if err != nil {
			t.Fatal(err)
		}

		txlist = append(txlist, tx)
		txhash := tx.Hash()
		txaddr[txhash] = addr
		txowner[txhash] = owner
	}

	txsresult := wolk[node].SendRawTransactions2(txlist)
	for txhash, err := range txsresult {
		if err != nil {
			log.Trace(fmt.Sprintf("Transaction error: tx(%x)  [Error%s]", txhash, err.Error()))
		}
	}

	for txhash, addr := range txaddr {
		owner := txowner[txhash]
		// initialize a map of which nodes have the tx
		hastx := make(map[int]bool)
		for i := 0; i < MAX_VALIDATORS; i++ {
			hastx[i] = false
		}
		start := time.Now()
		for done := false; !done; {
			count := 0
			for i := 0; i < MAX_VALIDATORS; i++ {
				if hastx[i] == false {
					_, bn, _, ok, err := wolk[i].GetTransaction(context.TODO(), txhash)
					if err != nil || !ok {
						log.Trace(fmt.Sprintf("Not included yet: tx(%x)  [Node%d] %s", txhash, i, time.Since(start)))
					} else if ok {
						log.Info(fmt.Sprintf("Included: tx(%x) in BLOCK %d [Node%d] %s", txhash, bn, i, time.Since(start)))
						hastx[i] = true
					}
				} else {
					count++
				}
			}
			log.Info(fmt.Sprintf("%d/%d nodes have SetName transaction for %s [%s]\n", count, MAX_VALIDATORS, owner, time.Since(start)))
			if count == MAX_VALIDATORS {
				done = true
			}
		}

		// do a getname call and see if it matches addr
		count := 0
		for i := 0; i < MAX_VALIDATORS; i++ {
			var options RequestOptions
			receivedaddr, ok, _, err := wolk[i].GetName(owner, &options)
			if err != nil {
				t.Fatalf("GetName ERR %v", err)
			} else if ok {
				log.Info("GetName Success", "node", i, "name", owner)
			} else {
				log.Info("GetName NOT FOUND", "node", i, "name", owner)
			}
			if bytes.Compare(addr.Bytes(), receivedaddr.Bytes()) != 0 {
				t.Fatalf("GetName INCORRECT addr %v != receivedaddr %v", addr, receivedaddr)
			} else {
				count++
			}
		}
		if count != MAX_VALIDATORS {
			t.Fatalf("INCONSISTENCY ACROSS nodes!")
		}
	}

	// TEST 1: addr1 is transferring 15 WOLK to addr0
	// CHECK:
	//  1. balance of address 0 should be 15 more
	//  2. balance of address 1 should 15 less!

	txlist = nil

	loop := 50
	amount := uint64(1)
	txs := make([]*Transaction, MAX_VALIDATORS*loop)
	balance_start := make([]uint64, MAX_VALIDATORS)
	addr0 := crypto.GetECDSAAddress(privateKeys[0])
	balance_start[0], _, _ = wolk[0].GetBalance(addr0, 0)
	j := 1
	for z := 1; z <= loop; z++ {
		for r := 1; r < MAX_VALIDATORS; r++ {

			addr := crypto.GetECDSAAddress(privateKeys[r])
			balance_start[r], _, _ = wolk[r].GetBalance(addr, 0)
			tx, err := NewTransaction(privateKeys[0], http.MethodPost, path.Join(ProtocolName, "transfer"), NewTxTransfer(amount, fmt.Sprintf("owner%d", r)))
			if err != nil {
				t.Fatalf("NewTransaction %v", err)
			}

			txlist = append(txlist, tx)
			txhash := tx.Hash()

			//// submit the Transfer transaction
			//txhash, err := wolk[(r+4)%MAX_VALIDATORS].SendRawTransaction(tx)
			//if err != nil {
			//	t.Fatal(err)
			//}
			txs[j] = tx
			j = j + 1
			log.Info("**** SENT TRANSFER TX ****", "txhash", txhash, "tx", tx)
		}
	}
	txsresult = wolk[node].SendRawTransactions2(txlist)
	for txhash, err := range txsresult {
		if err != nil {
			log.Trace(fmt.Sprintf("Transaction error: tx(%x)  [Error%s]", txhash, err.Error()))
		}
	}

	for r := 1; r < (MAX_VALIDATORS*loop - loop); r++ {
		tx := txs[r]
		hastx := make(map[int]bool)
		for i := 0; i < MAX_VALIDATORS; i++ {
			hastx[i] = false
		}
		start := time.Now()
		for done := false; !done; {
			missing := 0
			for i := 0; i < MAX_VALIDATORS; i++ {
				if hastx[i] == false {
					_, bn, _, ok, err := wolk[i].GetTransaction(context.TODO(), tx.Hash())
					if err != nil || !ok {
						log.Trace(fmt.Sprintf("Not included yet: tx(%x)  [Node%d] %s", tx.Hash(), i, time.Since(start)))
						missing++
					} else if ok {
						log.Trace(fmt.Sprintf("Included: tx(%x) in BLOCK %d [Node%d] %s", tx.Hash(), bn, i, time.Since(start)))
						hastx[i] = true
					}
				}
			}
			log.Info(fmt.Sprintf("%d/%d nodes are *missing* transfer transaction %d [%s]\n", missing, (MAX_VALIDATORS*loop - loop), r, time.Since(start)))
			if missing == 0 {
				done = true
			} else {
				time.Sleep(1000 * time.Millisecond)
			}
		}
	}

	for i := 0; i < MAX_VALIDATORS; i++ {
		// do GetBalance calls and see if it matches
		balance0_end, ok, err := wolk[i].GetBalance(addr0, 0)
		if err != nil {
			t.Fatalf("GetBalance0 ERR %v", err)
		} else if ok {
			log.Info("GetBalance0 Success", "node", i, "addr", addr0, "balance", balance0_end)
		} else {
			log.Info("GetBalance NOT FOUND", "node", i, "addr", addr0)
		}
		expected_diff := int64(-amount) * int64(MAX_VALIDATORS*loop-loop)
		diff := int64(balance0_end) - int64(balance_start[0])
		if diff != expected_diff {
			t.Fatalf("TX TEST Balance0 FAILURE %d != %d", diff, expected_diff)
		} else {
			log.Info("TX TEST Balance0 SUCCESS")
		}

		count := 0
		for r := 1; r < MAX_VALIDATORS; r++ {
			addr := crypto.GetECDSAAddress(privateKeys[r])
			balance_end, ok1, err := wolk[r].GetBalance(addr, 0)
			if err != nil {
				t.Fatalf("GetBalance1 ERR %v", err)
			} else if !ok1 {
				log.Info("GetBalance1 NOT FOUND", "node", i, "addr", addr)
			}
			diff := int64(balance_end) - int64(balance_start[r])
			if diff != int64(int64(amount)*int64(loop)) {
				t.Fatalf("TX TEST Node%d Balance%d FAILURE %d != %d", i, r, diff, amount)
			} else {
				log.Info("TX TEST GetBalance SUCCESS", "Node", i, "Balance", balance_end, "addr", addr)
				count++
			}
		}
		if count != MAX_VALIDATORS-1 {
			t.Fatalf("INCONSISTENCY ACROSS nodes!")
		}
		log.Info("TX TEST CONSISTENT", "Node", i)
	}
}
