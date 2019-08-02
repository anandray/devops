package wolk

import (
	"time"
)

var (
	UserAmount   uint64 = 8
	TokenPerUser uint64 = 128

	TotalTokenAmount uint64 = 8 * 128 // UserAmount * TokenPerUser

	Malicious      uint64 = 0
	IsVoting       uint64 = 1
	NetworkLatency uint64 = 0
)

func setConsensusParameters(registrySize int) {
	MULT = registrySize / 8
	if MULT < 1 {
		MULT = 1
	}
	expectedSubusers = 100 * MULT
}

var (
	MULT             = 1
	expectedSubusers = 100 // this is in token space  ==> 10% of 125*8 when MULT=1; 10% of 125*32 when MULT=4
)

const (
	LastConsensusState = 0
	LastExternalState  = -1
	PreemptiveState    = -2
	LastFinalizedState = -3
	LocalBestState     = -4
)

const (
	expectedTokensProposer = 100
	expectedTokensVoter    = 100

	//64 node
	expectedTokensTentative = 63
	expectedTokensFinal     = 148

	thresholdFinal     = 0.74
	thresholdTentative = 0.50

	lambdaStep      = 60 * time.Second // λ_Step 			timeout for making a real step
	lambdaHeartbeat = 1 * time.Second  // λ_Heartbeat 			timeout for repeating vote

	maxTrailingBlocks       = 10 // max of 10 tentative blocks before txn cap
	tentativeTxnCap         = 50 // txn cap after exceeding maxTrailingBlocks
	maxTransactionsProposal = 384

	R = 1000 // seed refresh interval (# of rounds)

	// helper const var
	roleVoter    = "voter"
	roleProposer = "proposer"

	// Malicious type
	Honest = iota
	EvilBlockProposal
	EvilVoteEmpty
	EvilVoteNothing
)
