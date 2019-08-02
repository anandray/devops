.PHONY: wolk wcloud wb wolktest cloudstore test Account Proposal Luby HTTP RPC Blockchain QBlock ReedSolomon Chunker Feistel Genesis Filetype CheckProof RSACoding NodeList Sequence SMT StorageEncode Storage Transaction

GOBIN = $(shell pwd)/build/bin
GO ?= latest

wolk:
		build/env.sh go run build/ci.go install ./cmd/wolk
		@echo "Done building wolk.  Run \"$(GOBIN)/wolk\" to launch wolk."

wcloud:
		go build -o build/bin/wcloud cmd/wcloud/wcloud.go
		@echo "Done building wcloud."
		@echo "Run \"$(GOBIN)/wcloud\" to launch wcloud."

chunkbench:
		@echo "chunkbench (Chunk benchmark)"
		go build -o build/bin/chunkbench cmd/chunkbench/chunkbench.go
		@echo "Done building chunkbench."

wolkbench:
		@echo "wolkbench (Wolk benchmark)"
		go build -o build/bin/wolkbench cmd/wolkbench/wolkbench.go
		@echo "Run \"$(GOBIN)/wolkbench\" to launch wolkbench."

shimserver:
		@echo "shimserver (Wolk Shim Server)"
		go build -o build/bin/shimserver cmd/shimserver/shimserver.go
		@echo "Run \"$(GOBIN)/shimserver\" to launch shimserver."

wolktest:
		@echo "Wolk Test"
		go test -v -short=1 ./wolk
#		go test -v -short=1 ./wolk

#	go test -v -short=1 -run QBlock ./wolk
#	go test -v -short=1 -run Production ./wolk
#	go test -v -short=1 -run Genesis  ./wolk
#	go test -short=1 -run ListenAndServe ./wolk
#	go test -short=1 -run SSLCheck ./wolk
#	go test -v -short=1 -run Filetype ./wolk
#	go test -v -short=1 -run Policy ./wolk
#	go test -v -short=1 -run InternalTx ./wolk
#	go test -v -short=1 -run SignTx ./wolk
#	go test -v -short=1 -run Transaction ./wolk
#	go test -v -short=1 -run ComputeDefaultHashes ./wolk
#	rm -rf ./wolk/tests/wvmtest.db
#	go test -v -short=1 -run SetWVMSchema ./wolk
#	go test -v -short=1 -run GetWVMSchema ./wolk
#	go test -v -short=1 -run DeleteLocalSchema ./wolk
#	go test -v -short=1 -run SetupSchema ./wolk/wvmc
#	go test -v -short=1 -run GetMasterTable ./wolk/wvmc
#	go test -v -short=1 -run ClearSchema ./wolk/wvmc
#	go test -short=1 -run WVMC ./wolk/wvmc
#	go test -short=1 -run SVM ./wolk
#
wolktestall:
	@echo "Wolk Test All"
	 	 -go test -v -run Account ./wolk
	 	 -go test -v -run Proposal ./wolk
	 	 -go test -v -run Luby ./wolk
	 	 -go test -v -run HTTP ./wolk
	 	 -go test -v -run QBlock ./wolk
	 	 -go test -v -run ReedSolomon ./wolk
	 	 -go test -v -run Chunker ./wolk
	 	 -go test -v -run Feistel ./wolk
	 	 -go test -v -run Genesis ./wolk
	 	 -go test -v -run Filetype ./wolk
	 	 -go test -v -run CheckProof ./wolk
	 	 -go test -v -run RSACoding ./wolk
	 	 -go test -v -run NodeList ./wolk
	 	 -go test -v -run Sequence ./wolk
	 	 -go test -v -run SMT ./wolk
	 	 -go test -v -run StorageEncode ./wolk
	 	 -go test -v -run Storage ./wolk
	 	 -go test -v -run Transaction ./wolk

cloudtest:
	@echo "Cloud Test"
		-go test -v -short=1 ./wolk/cloud
cloudtestall:
	@echo "Cloud Test All"
	 	 -go test -v -run GoogleDatastore ./wolk/cloud
	 	 -go test -v -run GoogleBigTable ./wolk/cloud
	 	 -go test -v -run AmazonDynamo ./wolk/cloud
	 	 -go test -v -run AlibabaTablestore ./wolk/cloud
	 	 -go test -v -run LevelDB ./wolk/cloud

GoogleDatastore:
	@echo "Test GoogleDatastore."
	-go test -v -run GoogleDatastore ./wolk/cloud

GoogleBigTable:
	@echo "Test GoogleBigTable."
	-go test -v -run GoogleBigTable ./wolk/cloud

AmazonDynamo:
	@echo "Test AmazonDynamo."
	-go test -v -run AmazonDynamo ./wolk/cloud

AlibabaTablestore:
	@echo "Test AlibabaTablestore."
	-go test -v -run AlibabaTablestore ./wolk/cloud

LevelDB:
	@echo "Test LevelDB."
	-go test -v -run LevelDB ./wolk/cloud

Account:
	 @echo "Test Account."
	 -go test -v -run Account ./wolk

Proposal:
	 @echo "Test Proposal."
	 -go test -v -run Proposal ./wolk

Luby:
	 @echo "Test Luby."
	 -go test -v -run Luby ./wolk

HTTP:
	 @echo "Test HTTP."
	 -go test -v -run HTTP ./wolk

RPC:
	 @echo "Test RPC."
	 -go test -v -run RPC ./wolk

Blockchain:
	 @echo "Test Blockchain."
	 -go test -v -run Blockchain ./wolk

QBlock:
	 @echo "Test QBlock."
	 -go test -v -run QBlock ./wolk

ReedSolomon:
	 @echo "Test ReedSolomon."
	 -go test -v -run ReedSolomon ./wolk

Chunker:
	 @echo "Test Chunker."
	 -go test -v -run Chunker ./wolk

Feistel:
	 @echo "Test Feistel."
	 -go test -v -run Feistel ./wolk

Genesis:
	 @echo "Test Genesis."
	 -go test -v -run Genesis ./wolk

Filetype:
	 @echo "Test Filetype."
	 -go test -v -run Filetype ./wolk

CheckProof:
	 @echo "Test CheckProof."
	 -go test -v -run CheckProof ./wolk

RSACoding:
	 @echo "Test RSACoding."
	 -go test -v -run RSACoding ./wolk

NodeList:
	 @echo "Test NodeList."
	 -go test -v -run NodeList ./wolk

Sequence:
	 @echo "Test Sequence."
	 -go test -v -run Sequence ./wolk

SMT:
	 @echo "Test SMT."
	 -go test -v -run SMT ./wolk

StorageEncode:
	 @echo "Test StorageEncode."
	 -go test -v -run StorageEncode ./wolk

Storage:
	 @echo "Test Storage."
	 -go test -v -run Storage ./wolk

Transaction:
	 @echo "Test Transaction."
	 -go test -v -run Transaction ./wolk
