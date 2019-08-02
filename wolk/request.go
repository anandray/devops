package wolk

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
)

const (
	DefaultMaxContentLength = 20000000
	DefaultShard            = 0
	DefaultSigType          = "ed25519-96"
)

type RequestOptions struct {
	Blockchain       string           `json:"blockchain,omitempty"`
	BlockNumber      int              `json:"blockNumber,omitempty"`
	Shard            uint64           `json:"shard,omitempty"`
	Sig              []byte           `json:"sig,omitempty"`
	SigType          string           `json:"sigType,omitempty"`
	MaxContentLength uint64           `json:"maxContentLength,omitempty"`
	Proof            bool             `json:"proof,omitempty"`
	ReplicaChallenge int64            `json:"replicachallenge,omitempty"`
	WaitForTx        string           `json:"waitfortx,omitempty"`
	IsPreemptive     bool             `json:"ispreemptive,omitempty"`
	History          string           `json:"history,omitempty"`
	Encryption       string           `json:"encryption,omitempty"`
	Range            string           `json:"range,omitempty"`
	Schema           string           `json:"schema,omitempty"`
	RequesterPays    uint8            `json:"requesterPays,omitempty"`
	ValidWriters     []common.Address `json:"validWriters,omitempty"`
	ShimURL          string           `json:"shimurl,omitempty"`

	PrimaryKey string         `json:"primaryKey,omitempty"` // AVL only
	Indexes    []*BucketIndex `json:"indexes,omitempty"`    // AVL only

	RangeField string `json:"rangeField,omitempty"`
	RangeStart string `json:"rangeStart,omitempty"`
	RangeEnd   string `json:"rangeEnd,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

func NewRequestOptions() *RequestOptions {
	var r RequestOptions
	r.Shard = DefaultShard
	r.MaxContentLength = DefaultMaxContentLength
	r.SigType = DefaultSigType
	return &r
}

func (s *RequestOptions) WithReplicaChallenge() bool {
	if s.ReplicaChallenge >= 0 {
		return true
	}
	return false
}

func (s *RequestOptions) WithProof() bool {
	if s.Proof == true {
		return true
	}
	return false
}

func (s *RequestOptions) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}
