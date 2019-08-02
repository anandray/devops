// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/log"
)

func TestStateObject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	log.New(log.LvlTrace, "", fmt.Sprintf("wolk-trace9"))
	wolk, _, _, _, nodelist := newWolkPeer(t, defaultConsensus)
	defer ReleaseNodes(t, nodelist)
	time.Sleep(1 * time.Second)
	ws := wolk[0]
	ctx := context.TODO()
	sdb, err := NewStateDB(ctx, ws.Storage, ws.LastKnownBlock().Hash())
	if err != nil {
		t.Fatalf("newStateDB %v", err)
	}
	rawKey := []byte("randomkey")
	rawVal := []byte("randomval")
	keyHash := common.BytesToHash(wolkcommon.Computehash(rawKey))
	valHash := common.BytesToHash(wolkcommon.Computehash(rawVal))
	so := NewStateObject(sdb, keyHash, valHash)

	if so.Key() != keyHash {
		t.Fatalf("expected key(%v) != actual key(%v)", keyHash, so.Key())
	}
	if so.Val() != valHash {
		t.Fatalf("expected val(%v) != actual val(%v)", valHash, so.Val())
	}
	alteredVal := []byte("alteredval")
	alteredValHash := common.BytesToHash(wolkcommon.Computehash(alteredVal))
	so.SetOwner(alteredValHash)
	if so.Val() != alteredValHash {
		t.Fatalf("expected altered val(%v) != actual altered val(%v)", alteredValHash, so.Val())
	}

	tprint("passed")
}
