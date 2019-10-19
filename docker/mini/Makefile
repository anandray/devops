.PHONY: wolk wcloud wolktest

GOBIN = $(shell pwd)/build/bin
GO ?= latest

T=$(shell date +%Y%m%d%H%M)

dmain:
		go build 
		docker build  --force-rm  -t gcr.io/us-west1-wlk/wolkinc/mini:$T .
		gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/mini:$T
		echo $T

