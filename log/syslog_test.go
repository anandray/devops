package log_test

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"testing"

	"github.com/wolkdb/cloudstore/log"
)

func TestLogDebug(t *testing.T) {
	thisOS := runtime.GOOS
	fmt.Printf("%s\n", thisOS)
	log.SysLog("hi", "wolk-debug")
}

func TestLogChunk(t *testing.T) {
	chunkID, _ := hex.DecodeString("6a9e288984801f51a1ff197805d15de3c19663cfd1a290860823ba8030313162")
	log.Chunk(chunkID, "PutChunk")
}
