package log

import (
	"log/syslog"
	"runtime"

	log "github.com/ethereum/go-ethereum/log"
)

const DEBUG = "wolk-debug"
const CLOUD = "wolk-cloud"
const CHUNK = "wolk-netstats"
const MINING = "wolk-mining"

type WolkCloudLog struct {
	// used in SetChunk/SetShare/GetChunk/GetShare/SetChunkBatch/GetChunkBatch
	Chain     string `json:"chain"`
	Chunks    uint64 `json:"chunks,omitempty"`
	Label     string `json:"label,omitempty"`
	Host      string `json:"host,omitempty"`
	Provider  string `json:"provider,omitempty"`
	Timestamp uint64 `json:"timestamp,omitempty"`
	Duration  uint64
	Extra     string `json:"extra,omitempty"`
}

/* Example Usage:

logLineObject := &WolkCloudLog{
       Chain:   "sql",
       ChainId: 77,
       SetChunk: 1,
       Label: "cloudster.SetChunk",
       Host: 127.0.0.1,
       ChunkId: "91dae949703266bcd67ef2dfa9a20d7b62b25520a8375c6a5fad16bd72ae9d4e",
       Bytes: 4096,
       Timestamp: uint64(time.Now().UnixNano()),
}

logLineString, _ := json.Marshal(logLineObject)
wolklog.SysLog(string(logLineString), wolklog.CLOUD)

*/

func SysLog(info string, path string) error {

	thisOS := runtime.GOOS
	if thisOS != "linux" {
		return nil
	}

	//path = wolk-cloud for timing lines
	//path = wolk-debug for debug lines
	logger, err := syslog.Dial("tcp", "127.0.0.1:5000", syslog.LOG_ERR, path) // connection to a log daemon

	if err != nil {
		log.Trace("[handler.go SysLog] syslog.Dial", "Error", err)
		return err
	} else {
		defer logger.Close()
		syslogerr := logger.Info(info)
		if syslogerr != nil {
			log.Trace("[handler.go SysLog] logger.Info", "Error", syslogerr)
		}
	}
	return nil
}
