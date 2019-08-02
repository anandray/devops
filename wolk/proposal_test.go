// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"fmt"
	"testing"

	wolkcommon "github.com/wolkdb/cloudstore/common"
	"github.com/wolkdb/cloudstore/crypto"
)

/*
package main

import "fmt"
import "math/big"
import "math"

func main() {

	n := int64(8)
	k := int64(5)
	sum := new(big.Float)
	p := .125
	for i:=int64(0); i<=k; i++ {
		i_f := float64(i)
		dsum := math.Pow(p, i_f) * math.Pow((1-p),float64(n-i))
		x := new(big.Float).SetInt(new(big.Int).Binomial(n, i))
		x.Mul(x, big.NewFloat(dsum))
		sum.Add(x, sum)
	}
	fmt.Printf("%s", sum)
}
*/

func TestSortition(t *testing.T) {
	seed := wolkcommon.Computehash([]byte(""))
	nrounds := uint64(10)

	for round := uint64(2); round < nrounds; round++ {
		proposers := 0
		voters := 0
		totalProposerSubusers := 0
		totalVoterSubusers := 0

		for id := uint64(0); id < uint64(UserAmount); id++ {
			privateStr := fmt.Sprintf("%x", wolkcommon.Computehash([]byte(fmt.Sprintf("%d", id))))
			privkey, _ := crypto.HexToEd25519(privateStr)

			proposerVRF, proposerProof, proposerSubusers := sortition(privkey, seed, role(roleProposer, round), expectedTokensProposer, TokenPerUser, TotalTokenAmount)
			if proposerSubusers > 0 {
				var newBlk Block
				newBlk.BlockNumber = round
				newBlk.Seed = seed
				proposal := &Proposal{
					Proposer:  id,
					Prior:     maxPriority(proposerVRF, proposerSubusers),
					VRF:       proposerVRF,
					Proof:     proposerProof,
					Signature: make([]byte, 65),
				}
				proposal.Sign(privkey)

				subusers2 := verifySort(privkey.PublicKey(), proposerVRF, proposerProof, seed, role(roleProposer, round), round, expectedTokensProposer, TokenPerUser, TotalTokenAmount)
				if subusers2 != proposerSubusers {
					t.Fatalf("verifySort failure: proposer")
				}
				proposers++
				totalProposerSubusers += subusers2
			}

			voterVRF, voterProof, voterSubusers := sortition(privkey, seed, role(roleVoter, round), expectedTokensVoter, TokenPerUser, TotalTokenAmount)
			if voterSubusers > 0 {
				voters++
				subusers1 := verifySort(privkey.PublicKey(), voterVRF, voterProof, seed, role(roleVoter, round), round, expectedTokensVoter, TokenPerUser, TotalTokenAmount)
				if subusers1 != voterSubusers {
					t.Fatalf("verifySort failure: voter")
				}
				totalVoterSubusers += subusers1
			}
			fmt.Printf("  %d\tvoterSubusers=%d\tproposerSubusers=%d\n", id, voterSubusers, proposerSubusers)
		}
		fmt.Printf("Round %03d|proposers=%03d|totalProposerSubusers=%03d|voters=%d|totalVoterSubusers=%03d\n", round, proposers, totalProposerSubusers, voters, totalVoterSubusers)

		seed = wolkcommon.Computehash(seed)
	}
}
