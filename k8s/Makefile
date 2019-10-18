.PHONY: wolk wcloud wolktest

GOBIN = $(shell pwd)/build/bin
GO ?= latest

T=$(shell date +%Y%m%d%H%M)

stage1:
		setupkube 1
		docker build  --force-rm  -t gcr.io/us-west1-wlk/wolkinc/cloudstore1:$T .
		gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/cloudstore1:$T
		pushkube 1 $T

stage2:
		setupkube 2
		docker build  --force-rm  -t gcr.io/us-west1-wlk/wolkinc/cloudstore2:$T .
		gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/cloudstore2:$T
		pushkube 2 $T

stage3:
		setupkube 3
		docker build  --force-rm  -t gcr.io/us-west1-wlk/wolkinc/cloudstore3:$T .
		gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/cloudstore3:$T
		pushkube 3 $T

stage4:
		setupkube 4
		docker build  --force-rm  -t gcr.io/us-west1-wlk/wolkinc/cloudstore4:$T .
		gcloud docker -- push gcr.io/us-west1-wlk/wolkinc/cloudstore4:$T
		pushkube 4 $T
