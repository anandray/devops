package wolk

import (
	"testing"
)

func TestNode(t *testing.T) {
	/*
		// TEST 2: node 2 is being purchased for 25000 WOLK by address 1
		// CHECK:
		//  1. balance of address of node 2 should be min{25000, vMin} more
		//  2. balance of address 1 should be min{25000, vMin} less
		balance1_start, _, _ = wolk[0].GetBalance(addr1, bn)
		balance2_start, _, _ := wolk[0].GetBalance(addr2, bn)
		node := uint64(2)
		ip := "127.0.0.1"
		consensusip := "127.0.0.1"
		value := uint64(TestBalance)
		region := uint8(1)
		tx, err = NewTransaction(privateKeys[1], http.MethodPut, path.Join(ProtocolName, "node", fmt.Sprintf("%d", node)), NewTxNode(ip, consensusip, region, value))
		if err != nil {
			t.Fatalf("NewTransaction %v", err)
		}
		log.Info("TX TEST 2 START", "tx", tx)
		doTransaction(t, wolk[0], tx)
		bn = wolk[0].blockNumber - 1
		balance1_end, _, _ = wolk[0].GetBalance(addr1, bn)
		balance2_end, _, _ := wolk[0].GetBalance(addr2, bn)
		diff2 := int64(balance2_end) - int64(balance2_start)
		diff1 = int64(balance1_end) - int64(balance1_start)
		log.Info("Addr", "addr1 (buyer)", addr0.Hex(), "addr2 (prev nodeOwner)", addr2.Hex())
		if diff1+diff2 == 0 && (diff1 < 0) && (diff2 > 0) {
			log.Info("TX TEST 2", "balance1_start", balance1_start, "balance2_start", balance2_start, "balance1_end", balance1_end, "balance2_end", balance2_end, "diff1", diff1, "diff2", diff2)
			log.Info("TX TEST 2 SUCCESS")
		} else {
			log.Info("TX TEST 2 FINISH", "balance1_start", balance1_start, "balance2_start", balance2_start, "balance1_end", balance1_end, "balance2_end", balance2_end, "diff1", diff1, "diff2", diff2)
			t.Fatalf("TX TEST 2 FAILURE")
		}

		// TEST 3: node 2 is being updated by address 1
		// CHECK:
		//  1. GetNode(2) should be updated with all the values!
		node = uint64(2)
		options := NewRequestOptions()
		n2_start, _, _, _ := wolk[0].GetNode(node, options)
		ip = "127.0.0.1"
		consensusip = "127.0.0.1"
		value = uint64(32272)
		region = uint8(4)
		tx, err = NewTransaction(privateKeys[1], http.MethodPut, path.Join(ProtocolName, "node", fmt.Sprintf("%d", node)), NewTxNode(ip, consensusip, region, value))
		if err != nil {
			t.Fatalf("NewTransaction %v", err)
		}
		log.Info("TX TEST 3 START", "tx", tx)
		doTransaction(t, wolk[0], tx)
		bn = wolk[0].blockNumber - 1
		n2_end, _, _, _ := wolk[0].GetNode(node, options)
		log.Info("TX TEST 3 FINISH", "tx", tx)
		if n2_start == nil || n2_end == nil {
			t.Fatalf("getnode failure")
		}
		if (n2_end.valueInt != value) || (n2_end.region != region) {
			log.Info("TX TEST 3 FINISH", "n2_start", n2_start.String(), "n2_end", n2_end.String())
			t.Fatalf("TX TEST 3 FAILURE")
		} else {
			log.Info("TX TEST 3", "n2_start", n2_start.String(), "n2_end", n2_end.String())
			log.Info("TX TEST 3 SUCCESS")
		}

		// TEST 4: nodes are storing chunkID, node 1 is making a claim that it is storing this chunkID
		// CHECK:
		//  DOESN'T WORK node 1 gets mining rewards for a validclaim
			bn = wolk[0].blockNumber - 1
			balance0_start, _, _ = wolk[0].GetBalance(addr0, bn)
			balance1_start, _, _ = wolk[0].GetBalance(addr1, bn)
			_, chunk := generateRandomData(chunkSize)
			_, err = wolk[0].SetChunk(nil, chunk)
			if err != nil {
				t.Fatalf("SetChunk %v", err)
			}

			tx, err = NewTransaction(privateKeys[1], http.MethodPatch, path.Join(ProtocolName, "account"), &TxBucket{Quota: MinimumQuota})
			if err != nil {
				t.Fatalf("NewTransaction %v", err)
			}
			doTransaction(t, wolk[0], tx)
			bn = wolk[0].blockNumber - 1
			balance0_end, _, _ = wolk[0].GetBalance(addr0, bn)
			balance1_end, _, _ = wolk[0].GetBalance(addr1, bn)
			diff0 = int64(balance0_end) - int64(balance0_start)
			diff1 = int64(balance1_end) - int64(balance1_start)
			log.Info("TX TEST 4 FINISH", "balance0_start", balance0_start, "balance1_start", balance1_start, "balance0_end", balance0_start, "balance1_end", balance1_end, "diff0", diff0, "diff1", diff1)
	*/
}

// TestAuth from plasma nosql
/*
func randString(rnd *rand.Rand, bytelength int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, bytelength)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)

}

type ValidKeyVal struct {
	encKey  []byte
	encVal  []byte
	key     string
	val     string
	dataSet []byte
	dataGet []byte
	auth    common.Address // 20 bytes
	sigSet  []byte
	sigGet  []byte
	privKey []byte
}

func TestAuth(t *testing.T) {
	//logprint()

		ctx := &node.ServiceContext{}

		chain, err := New(ctx, &DefaultTestConfig)
		if err != nil {
			t.Fatal(err)
		}
		defer chain.Close()
		service := deep.NewCliqueService(ctx.EventMux, chain)
		//defer service.Stop()  // this causes a panic. why?
		minter := deep.NewMinterPOAStruct(nil, service, 0)

		// read genesis, see if auth res is done corrrectly
		auths := make(map[int]common.Address, len(chain.authAddr))
		i := 0
		for addrnum := range chain.authAddr {
			auths[i] = addrnum
			tprint("auth addrs: (%d) (%x)", i, auths[i])
			i++
		}
		if len(auths) == 0 {
			t.Fatalf("no authorized addresses found in genesis.json")
		}

		// set up some predetermined privatekey, address pairs so we don't depend on a specific genesis.json for this test
		auths = map[int]common.Address{
			0: common.HexToAddress("eD23E17404C49D5422768Cb81848C66a13d5206b"),
			1: common.HexToAddress("2d2Ab4cF5031E5F0a62206f18EfE22D8071618Fb"),
			2: common.HexToAddress("4bb2e070867F8C32C7266BADB1896DDAfeFA8533"),
		}
		privkeys := map[int][]byte{
			0: common.Hex2Bytes("a2059bef839e761f887a209f916b82394a9e755d274c11903bd019177caeb5ab"),
			1: common.Hex2Bytes("2d62d6c0d973011c823b006d6b621fc2ad2d51293e260c62aa138bd407f0a00d"),
			2: common.Hex2Bytes("194a4179320ab49c70afb299715d1ced4dd5320e8a0385ccc04116cbbc15aca6"),
		}
		// copy over the chain's authorized addresses for this test
		chain.authAddr = make(map[common.Address]int)
		for _, addr := range auths {
			chain.authAddr[addr] = 1
		}

		tprint("modified chain's authorized addresses to:")
		for a := range chain.authAddr {
			tprint("%x", a)
		}
		// cmp auth to setkey getkey addr
		numkeys := 5
		keylength := 16
		vallength := 16
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		//keyvals := make([]ValidKeyVal, numkeys)
		var keyvals []ValidKeyVal

		// generate data
		for i := 0; i < numkeys; i++ {
			key := randString(rnd, keylength)
			val := randString(rnd, vallength)
			authnum := rnd.Intn(len(chain.authAddr))
			kv := new(ValidKeyVal)
			kv.key = key
			kv.val = val
			kv.auth = auths[authnum]
			kv.privKey = privkeys[authnum]
			kv.encKey = wolkcommon.Computehash([]byte(key))
			kv.encVal = []byte(val)
			kv.dataSet = encodeDataForSignature(chain.blockchainID, kv.encKey, kv.encVal)
			var emptyVal []byte
			kv.dataGet = encodeDataForSignature(chain.blockchainID, kv.encKey, emptyVal)
			kv.sigSet, err = deep.Sign(kv.dataSet, kv.privKey)
			if err != nil {
				t.Fatal(err)
			}
			kv.sigGet, err = deep.Sign(kv.dataGet, kv.privKey)
			if err != nil {
				t.Fatal(err)
			}
			if err != nil {
				t.Fatal(err)
			}
			keyvals = append(keyvals, *kv)
		}

		// set keys with wrong sig
		for _, kv := range keyvals {
			badprivkey, _, err := deep.GenerateAuthKeys()
			if err != nil {
				t.Fatal(err)
			}
			badsig, err := deep.Sign(kv.dataSet, badprivkey)
			if err != nil {
				t.Fatal(err)
			}
			txhash, err := chain.SetKey(chain.blockchainID, badsig, kv.encKey, kv.encVal)
			if err == nil {
				t.Fatalf("SetKey should not have allowed bad sig(%x) to pass", badsig)
			}
			tprint("SetKey failed txn: key(%s) val(%s) badsig(%x) txhash(%x)", kv.key, kv.val, badsig, txhash)
		}

		// set keys
		for _, kv := range keyvals {
			txhash, err := chain.SetKey(chain.blockchainID, kv.sigSet, kv.encKey, kv.encVal)
			if err != nil {
				t.Fatal(err)
			}
			tprint("SetKey txn: key(%s) val(%s) sig(%x) txhash(%x)", kv.key, kv.val, kv.sigSet, txhash)
		}

		blocknum := uint64(chain.LastKnownBlock().Number())
		tprint("Current Blocknumber: %v", blocknum)

		minter.MintNewBlock()

		blocknum = uint64(chain.LastKnownBlock().Number())
		tprint("Minted new block. Current Blocknumber: %v", blocknum)

		// get keys with wrong auth: commented out b/c getkey sig is disabled
		// for _, kv := range keyvals {
		// 	badprivkey, _, err := deep.GenerateAuthKeys()
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}
		// 	badsig, err := deep.Sign(kv.dataGet, badprivkey)
		// 	if err != nil {
		// 		t.Fatal(err)
		// 	}
		// 	proof, _, err := chain.GetKey(chain.blockchainID, badsig, kv.encKey, blocknum)
		// 	if err == nil {
		// 		t.Fatalf("GetKey should not have allowed (%x) to pass", badsig)
		// 	}
		// 	tprint("GetKey failed txn: key(%s) val(%s) sig(%x) proof(%x)", kv.key, kv.val, badsig, proof)
		// }

		// get keys
		for _, kv := range keyvals {
			proof, encVal, err := chain.GetKey(chain.blockchainID, kv.sigGet, kv.encKey, uint64(blocknum))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encVal, kv.encVal) {
				t.Fatalf("GetKey (%s) got the wrong value (%x), should be (%x)", kv.key, encVal, kv.encVal)
			}
			hasproof := false
			if len(proof.String()) > 0 {
				hasproof = true
			}
			tprint("GetKey txn: key(%s) val(%s) sig(%x) hasproof(%v)", kv.key, kv.val, kv.sigGet, hasproof)
		}

}
*/
