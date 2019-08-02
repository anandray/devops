.PHONY:	wolktest transaction block config dbchunkstore server keymanager bplus hashdb

GOBIN = $(shell pwd)/build/bin
GO ?= latest

wolktest:
	@echo "Wolk Test"
	go test -v -short=1 -run Account
	go test -v -short=1 -run SubUsersSingle
	go test -v -short=1 -run Algorand
	go test -v -short=1 -run BandwidthCheck
	go test -v -short=1 -run QBlock
	go test -v -short=1 -run Production
	go test -v -short=1 -run Genesis 
	go test -short=1 -run ListenAndServe
	go test -short=1 -run SSLCheck
	go test -v -short=1 -run Filetype
	go test -v -short=1 -run Policy
	go test -v -short=1 -run InternalTx
	go test -v -short=1 -run SignTx
	go test -v -short=1 -run Transaction
	go test -v -short=1 -run ComputeDefaultHashes
	rm -rf ./tests/wvmtest.db
	go test -v -short=1 -run SetWVMSchema
	go test -v -short=1 -run GetWVMSchema
	go test -v -short=1 -run DeleteLocalSchema
	go test -v -short=1 -run SetupSchema/wvmc
	go test -v -short=1 -run GetMasterTable/wvmc
	go test -v -short=1 -run ClearSchema/wvmc
	go test -short=1 -run WVMC/wvmc
	go test -short=1 -run SVM

test:
	@echo "test all."
#	@echo "test config."
#	-go test -v ./config_test.go
#	@echo "test txpool"	
#	-go test -v ./txpool_test.go
#	@echo "test transaction."
#	-go test -v ./transaction_test.go
#	@echo "test block."
#	-go test -v ./block_test.go
#	@echo "test dbmanager."
#	-go test -v ./dbmanager_test.go
#	@echo "test bplus."
#	-go test -v ./bplus_test.go
#	@echo "test query."
#	-go test -v ./query_test.go
#	@echo "test wvm."
#	-go test -v ./wvm_test.go

#transaction:
#	@echo "test transaction."
#	go test -v ./transaction_test.go

#block:
#	@echo "test block."
#	go test -v ./block_test.go

config:
	@echo "test config."
	go test -v ./config_test.go

txpool:
	@echo "test txpool"
	go test -v ./txpool_test.go

#bplus:
#	@echo "test bplus."
#	go test -v bplus_test.go

#query:
#	@echo "test query."
#	go test -v ./query_test.go

#sqlchain:
#	@echo "test dbmanager."
#	go test -v ./dbmanager_test.go
