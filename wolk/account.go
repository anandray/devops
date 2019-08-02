// Copyright 2018 Wolk Inc.
// This file is part of the Wolk library.
package wolk

import (
	"crypto/rsa"
	"encoding/json"

	jose "gopkg.in/square/go-jose.v2"
)

type Account struct {
	Balance      uint64 `json:"balance,omitempty"`
	Quota        uint64 `json:"quota,omitempty"`
	Usage        uint64 `json:"usage,omitempty"`
	LastClaim    uint64 `json:"lastClaim,omitempty"`
	RSAPublicKey []byte `json:"rsaPublicKey,omitempty"`
	ShimURL      string `json:"shimURL,omitempty"` //if get/(post?) refers to bucket/collection not found this shimURL will be pre-pended to what follows after owner name to retrieve content
}

func NewAccount() *Account {
	return &Account{
		Balance:      TestBalance, // FOR TESTING
		Quota:        0,
		Usage:        0,
		LastClaim:    0,
		RSAPublicKey: []byte(""),
		ShimURL:      "",
	}
}

func (a *Account) String() string {
	bytes, err := json.Marshal(a)
	if err != nil {
		return "{}"
	} else {
		return string(bytes)
	}
}

func (a *Account) GetRSAPublicKey() (pk *rsa.PublicKey, err error) {
	var j jose.JSONWebKey
	err = j.UnmarshalJSON(a.RSAPublicKey)
	if err != nil {
		return pk, err
	}
	// TODO do some type checking
	rsaKey := j.Key.(*rsa.PublicKey)
	return rsaKey, nil
}
