// Copyright 2018 Wolk Inc.
// This file is part of the Wolk Deep Blockchains library.
package wolk

import (
	"context"
	"fmt"

	"github.com/wolkdb/cloudstore/log"
)

//SQL api -
func (wolkstore *WolkStore) Read(req *SQLRequest, options *RequestOptions) (result *SQLResponse, err error) {
	log.Info("[backend_sql:Read] start")
	result, _, err = wolkstore.read(req, options)
	log.Info("[backend_sql:Read] end")
	return result, err
}

//func (wolkstore *WolkStore) GetMutateTransaction(req *SQLRequest) (txp *SQLRequest, err error) {
//	return wolkstore.getMutateTransaction(req)
//}

//backend SQL functions
func (wolkstore *WolkStore) read(req *SQLRequest, options *RequestOptions) (result *SQLResponse, proof *Proof, err error) {

	log.Info("[backend_sql:read] start")
	if options == nil {
		log.Error("[backend_sql:read] empty options... why?")
		options = NewRequestOptions()
	}
	log.Info("[backend_sql:read]", "options", options.String())

	state, ok, err := wolkstore.Storage.getStateDB(context.TODO(), options.BlockNumber)
	if err != nil {
		log.Error("[backend_sql:read] GetBlockByNumber")
		return result, proof, fmt.Errorf("[backend_sql:read] %s", err)
	}
	if !ok {
		log.Error("[backend_sql:read] GetBlockByNumber NOT OK")
		return result, proof, fmt.Errorf("[backend:GetKey] no block")
	}

	shResult, proof, err := state.SelectHandler(req, options.WithProof())
	if err != nil {
		log.Error(fmt.Sprintf("[backend:SQLRead] Error in Select Handler | %+v", err))
		return result, proof, fmt.Errorf("[backend_sql:read] %s", err)
	}
	result = &shResult
	log.Info("[backend_sql:read] SUCCESS", "result", result)
	return result, proof, nil
}
